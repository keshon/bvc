package block

import (
	"app/internal/fs"
	"app/internal/util"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/zeebo/xxh3"
)

const (
	minChunkSize = 2 * 1024 * 1024 // 2 MiB
	maxChunkSize = 8 * 1024 * 1024 // 8 MiB
	rollMod      = 4096
	readBufSize  = 32 * 1024 // 32 KiB streaming read buffer
)

// BlockRef describes one physical block of content.
type BlockRef struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
}

// BlockStatus indicates the state of a block on disk.
type BlockStatus int

const (
	OK BlockStatus = iota
	Missing
	Damaged
)

// BlockCheck contains information about a single block.
type BlockCheck struct {
	Hash     string
	Status   BlockStatus
	Files    []string
	Branches []string
}

// BlockContext handles all object-level storage operations.
type BlockContext struct {
	BlocksDir string // path to the blocks root directory (.bvc/objects)
	FS        fs.FS  // block filesystem abstraction
}

// NewBlockContext creates a new BlockContext.
func NewBlockContext(root string, fs fs.FS) *BlockContext {
	return &BlockContext{BlocksDir: root, FS: fs}
}

func (bc *BlockContext) GetBlocksDir() string {
	return bc.BlocksDir
}

// Read retrieves a block by its hash.
func (bc *BlockContext) Read(hash string) ([]byte, error) {
	path := filepath.Join(bc.BlocksDir, hash+".bin")
	data, err := bc.FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read block %q: %w", hash, err)
	}
	return data, nil
}

// Write stores all blocks for a given file.
func (bc *BlockContext) Write(filePath string, blocks []BlockRef) error {
	if err := bc.FS.MkdirAll(bc.BlocksDir, 0o755); err != nil {
		return fmt.Errorf("create objects dir: %w", err)
	}
	workers := util.WorkerCount()
	return util.Parallel(blocks, workers, func(b BlockRef) error {
		return bc.writeBlockAtomic(filePath, b)
	})
}

// writeBlockAtomic writes a block to disk atomically.
func (bc *BlockContext) writeBlockAtomic(filePath string, block BlockRef) error {
	dst := filepath.Join(bc.BlocksDir, block.Hash+".bin")

	// Skip if block exists
	if fi, err := bc.FS.Stat(dst); err == nil && fi.Size() == block.Size {
		return nil
	}

	if err := bc.FS.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ensure dir for %q: %w", dst, err)
	}

	// Read block from source
	src, err := bc.FS.Open(filePath)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", filePath, err)
	}
	defer src.Close()

	if _, err := src.Seek(block.Offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek to offset %d in %q: %w", block.Offset, filePath, err)
	}

	blockData := make([]byte, block.Size)
	if _, err := io.ReadFull(src, blockData); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read block %q: %w", block.Hash, err)
	}

	// Write block atomically via FS abstraction
	tmp, tmpPath, err := bc.FS.CreateTempFile(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %q: %w", filepath.Dir(dst), err)
	}
	defer bc.FS.Remove(tmpPath)

	if _, err := tmp.Write(blockData); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp block: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp block: %w", err)
	}

	if err := bc.FS.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("rename temp %q to %q: %w", tmpPath, dst, err)
	}

	return nil
}

// CleanupTemp removes orphaned temp files from blocks root directory (.bvc/objects).
func (bc *BlockContext) CleanupTemp() error {
	entries, err := bc.FS.ReadDir(bc.BlocksDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Keep same prefix behavior you had; remove 0-sized or unreadable tmp files.
		if strings.HasPrefix(name, "tmp-") || strings.HasPrefix(name, ".tmp-") {
			p := filepath.Join(bc.BlocksDir, name)
			if fi, err := bc.FS.Stat(p); err != nil || fi.Size() == 0 {
				_ = bc.FS.Remove(p)
			}
		}
	}
	return nil
}

// Verify checks a set of block hashes concurrently and streams results.
// We reuse util.Parallel for the worker pool behavior. VerifyBlock maps errors
// into BlockStatus, so we intentionally ignore util.Parallel's error semantics
// (we return nil in the worker) to ensure the whole set is processed.
func (bc *BlockContext) Verify(hashes map[string]struct{}, workers int) <-chan BlockCheck {
	out := make(chan BlockCheck, 128)
	if workers <= 0 {
		workers = util.WorkerCount()
	}

	go func() {
		defer close(out)

		// Convert map to slice
		list := make([]string, 0, len(hashes))
		for h := range hashes {
			list = append(list, h)
		}

		// Use Parallel worker pool; workers send BlockCheck into out channel.
		_ = util.Parallel(list, workers, func(h string) error {
			status, _ := bc.VerifyBlock(h)
			out <- BlockCheck{Hash: h, Status: status}
			return nil
		})
	}()

	return out
}

// VerifyBlock checks a single block for integrity using the selected hash.
// Blocks are modestly sized (<= maxChunkSize), so reading into memory is fine.
func (bc *BlockContext) VerifyBlock(hash string) (BlockStatus, error) {
	path := filepath.Join(bc.BlocksDir, hash+".bin")
	data, err := bc.FS.ReadFile(path)
	if err != nil {
		if bc.FS.IsNotExist(err) {
			return Missing, nil
		}
		// Treat read errors as damaged block.
		return Damaged, err
	}

	h := xxh3.Hash128(data).Bytes()
	actual := hex.EncodeToString(h[:])

	if actual == hash {
		return OK, nil
	}
	return Damaged, nil
}

// SplitFile divides a file into content-defined blocks deterministically using a
// Gear-like rolling hash. The function streams the file and avoids huge
// allocations. It returns BlockRefs in the order found.
func (bc *BlockContext) SplitFile(path string) ([]BlockRef, error) {
	fi, err := bc.FS.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file %q: %w", path, err)
	}
	if fi.Size() == 0 {
		return nil, nil
	}

	f, err := bc.FS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}
	defer f.Close()

	var (
		allBlocks []BlockRef
		offset    int64
	)

	// streaming read buffer
	readBuf := make([]byte, readBufSize)

	// accumulating block buffer (grow up to maxChunkSize)
	blockBuf := make([]byte, 0, min(minChunkSize, 64*1024)) // start with small cap

	var rh uint32
	var blockSize int

	for {
		n, rerr := f.Read(readBuf)
		if n > 0 {
			data := readBuf[:n]
			for i := 0; i < len(data); i++ {
				b := data[i]
				// append to block buffer (grow as needed, bounded by maxChunkSize)
				blockBuf = append(blockBuf, b)
				blockSize++

				// Gear-like mixing: shift + table lookup
				rh = (rh << 1) + gearTable[b]

				// decide split
				if shouldSplitBlock(blockSize, rh) {
					// finalize block
					br := hashBlock(blockBuf, offset)
					allBlocks = append(allBlocks, br)
					offset += br.Size

					// reset block buffer & rolling hash
					blockBuf = blockBuf[:0]
					blockSize = 0
					rh = 0
				} else if blockSize >= maxChunkSize {
					// forced max split
					br := hashBlock(blockBuf, offset)
					allBlocks = append(allBlocks, br)
					offset += br.Size
					blockBuf = blockBuf[:0]
					blockSize = 0
					rh = 0
				}
			}
		}

		if rerr != nil {
			if rerr == io.EOF {
				break
			}
			return nil, fmt.Errorf("read file %q: %w", path, rerr)
		}
	}

	// flush remaining bytes
	if len(blockBuf) > 0 {
		br := hashBlock(blockBuf, offset)
		allBlocks = append(allBlocks, br)
	}

	return allBlocks, nil
}

func shouldSplitBlock(size int, rh uint32) bool {
	return (size >= minChunkSize && rh%rollMod == 0) || size >= maxChunkSize
}

// hashBlock computes the hash of data using xxh3-128 and returns a BlockRef.
// Note: data is copied by xxh3 hashing; caller may reuse underlying slice.
func hashBlock(data []byte, offset int64) BlockRef {
	h := xxh3.Hash128(data).Bytes()
	return BlockRef{
		Hash:   hex.EncodeToString(h[:]),
		Size:   int64(len(data)),
		Offset: offset,
	}
}

// simple helpers

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// gearTable: 256 random uint32 values for Gear-like rolling hash.
// These constants can be replaced with any 256 uniformly random uint32s.
var gearTable = [256]uint32{
	0x243F6A88, 0x85A308D3, 0x13198A2E, 0x03707344,
	0xA4093822, 0x299F31D0, 0x082EFA98, 0xEC4E6C89,
	0x452821E6, 0x38D01377, 0xBE5466CF, 0x34E90C6C,
	0xC0AC29B7, 0xC97C50DD, 0x3F84D5B5, 0xB5470917,
	0x9216D5D9, 0x8979FB1B, 0xD1310BA6, 0x98DFB5AC,
	0x2FFD72DB, 0xD01ADFB7, 0xB8E1AFED, 0x6A267E96,
	0xBA7C9045, 0xF12C7F99, 0x24A19947, 0xB3916CF7,
	0x0801F2E2, 0x858EFC16, 0x636920D8, 0x71574E69,
	0xA458FEA3, 0xF4933D7E, 0x0D95748F, 0x728EB658,
	0x718BCD58, 0x82154AEE, 0x7B54A41D, 0xC25A59B5,
	0x9C30D539, 0x2AF26013, 0xC5D1B023, 0x286085F0,
	0xCA417918, 0xB8DB38EF, 0x8E79DCB0, 0x603A180E,
	0x6C9E0E8B, 0xB01E8A3E, 0xD71577C1, 0xBD314B27,
	0x78AF2FDA, 0x55605C60, 0xE65525F3, 0xAA55AB94,
	0x57489862, 0x63E81440, 0x55CA396A, 0x2AAB10B6,
	0xB4CC5C34, 0x1141E8CE, 0xA15486AF, 0x7C72E993,
	0xB3EE1411, 0x636FBC2A, 0x2BA9C55D, 0x741831F6,
	0xCE5C3E16, 0x9B87931E, 0xAFD6BA33, 0x6C24CF5C,
	0x7A325381, 0x28958677, 0x3B8F4898, 0x6B4BB9AF,
	0xC4BFE81B, 0x66282193, 0x61D809CC, 0xFB21A991,
	0x487CAC60, 0x5DEC8032, 0xEF845D5D, 0xE98575B1,
	0xDC262302, 0xEB651B88, 0x23893E81, 0xD396ACC5,
	0x0F6D6FF3, 0x83F44239, 0x2E0B4482, 0xA4842004,
	0x69C8F04A, 0x9E1F9B5E, 0x21C66842, 0xF6E96C9A,
	0x670C9C61, 0xABD388F0, 0x6A51A0D2, 0xD8542F68,
	0x960FA728, 0xAB5133A3, 0x6EEF0B6C, 0x137A3BE4,
	0xBA3BF050, 0x7EFB2A98, 0xA1F1651D, 0x39AF0176,
	0x66CA593E, 0x82430E88, 0x8CEE8619, 0x456F9FB4,
	0x7D84A5C3, 0x3B8B5EBE, 0xE06F75D8, 0x85C12073,
	0x401A449F, 0x56C16AA6, 0x4ED3AA62, 0x363F7706,
	0x1BFEDF72, 0x429B023D, 0x37D0D724, 0xD00A1248,
	0xDB0FEAD3, 0x49F1C09B, 0x075372C9, 0x80991B7B,
	0x25D479D8, 0xF6E8DEF7, 0xE3FE501A, 0xB6794C3B,
	0x976CE0BD, 0x04C006BA, 0xC1A94FB6, 0x409F60C4,
	0x5E5C9EC2, 0x196A2463, 0x68FB6FAF, 0x3E6C53B5,
	0x1339B2EB, 0x3B52EC6F, 0x6DFC511F, 0x9B30952C,
	0xCC814544, 0xAF5EBD09, 0xBEE3D004, 0xDE334AFD,
	0x660F2807, 0x192E4BB3, 0xC0CBA857, 0x45C8740F,
	0xD20B5F39, 0xB9D3FBDB, 0x5579C0BD, 0x1A60320A,
	0xD6A100C6, 0x402C7279, 0x679F25FE, 0xFB1FA3CC,
	0x8EA5E9F8, 0xDB3222F8, 0x3C7516DF, 0xFD616B15,
	0x2F501EC8, 0xAD0552AB, 0x323DB5FA, 0xFD238760,
	0x53317B48, 0x3E00DF82, 0x9E5C57BB, 0xCA6F8CA0,
	0x1A87562E, 0xDF1769DB, 0xD542A8F6, 0x287EFFC3,
	0xAC6732C6, 0x8C4F5573, 0x695B27B0, 0xBBCA58C8,
	0xE1FFA35D, 0xB8F011A0, 0x10FA3D98, 0xFD2183B8,
	0x4AFCB56C, 0x2DD1D35B, 0x9A53E479, 0xB6F84565,
	0xD28E49BC, 0x4BFB9790, 0xE1DDF2DA, 0xA4CB7E33,
	0x62FB1341, 0xCEE4C6E8, 0xEF20CADA, 0x36774C01,
	0xD07E9EFE, 0x2BF11FB4, 0x95DBDA4D, 0xAE909198,
	0xEAAD8E71, 0x6B93D5A0, 0xD08ED1D0, 0xAFC725E0,
	0x8E3C5B2F, 0x8E7594B7, 0x8FF6E2FB, 0xF2122B64,
	0x8888B812, 0x900DF01C, 0x4FAD5EA0, 0x688FC31C,
	0xD1CFF191, 0xB3A8C1AD, 0x2F2F2218, 0xBE0E1777,
	0xEA752DFE, 0x8B021FA1, 0xE5A0CC0F, 0xB56F74E8,
	0x18ACF3D6, 0xCE89E299, 0xB4A84FE0, 0xFD13E0B7,
	0x7CC43B81, 0xD2ADA8D9, 0x165FA266, 0x80957705,
	0x93CC7314, 0x211A1477, 0xE6AD2065, 0x77B5FA86,
	0xC75442F5, 0xFB9D35CF, 0xEBCDAF0C, 0x7B3E89A0,
	0xD6411BD3, 0xAE1E7E49, 0x00250E2D, 0x2071B35E,
	0x226800BB, 0x57B8E0AF, 0x2464369B, 0xF009B91E,
}
