package block

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"app/internal/config"
	"app/internal/util"

	"github.com/zeebo/xxh3"
)

const (
	minChunkSize = 2 * 1024 * 1024 // 2 MiB
	maxChunkSize = 8 * 1024 * 1024 // 8 MiB
	rollMod      = 4096
)

type BlockRef struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
}

type BlockStatus int

const (
	OK BlockStatus = iota
	Missing
	Damaged
)

type BlockCheck struct {
	Hash     string
	Status   BlockStatus
	Files    []string
	Branches []string
}

// SplitFileIntoBlocks splits a file into content-defined blocks.
func SplitFileIntoBlocks(path string) ([]BlockRef, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		blocks []BlockRef
		chunk  []byte
		buf    = make([]byte, maxChunkSize)
		offset int64
		rh     uint32
	)

	for {
		n, err := f.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				b := buf[i]
				chunk = append(chunk, b)
				rh = (rh<<1 + uint32(b)) & 0xFFFFFFFF
				if shouldSplit(len(chunk), rh) {
					blocks = append(blocks, hash(chunk, offset))
					offset += int64(len(chunk))
					chunk = chunk[:0]
				}
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		if n == 0 {
			break
		}
	}

	if len(chunk) > 0 {
		blocks = append(blocks, hash(chunk, offset))
	}

	return blocks, nil
}

func shouldSplit(size int, rh uint32) bool {
	return (size >= minChunkSize && rh%rollMod == 0) || size >= maxChunkSize
}

func hash(data []byte, offset int64) BlockRef {
	hash := xxh3.Hash128(data).Bytes()
	return BlockRef{
		Hash:   fmt.Sprintf("%x", hash),
		Size:   int64(len(data)),
		Offset: offset,
	}
}

// Store writes all blocks concurrently and safely on Windows.
func Store(filePath string, blocks []BlockRef) error {
	workers := util.WorkerCount()
	return util.Parallel(blocks, workers, func(b BlockRef) error {
		return WriteAtomic(filePath, b)
	})
}

// writeAtomic streams a single block to a temp file, then renames it.
func WriteAtomic(filePath string, block BlockRef) error {
	dst := filepath.Join(config.ObjectsDir, block.Hash+".bin")
	if fi, err := os.Stat(dst); err == nil && fi.Size() == block.Size {
		return nil
	}

	src, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), "tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	// Stream copy block
	if _, err := src.Seek(block.Offset, io.SeekStart); err != nil {
		tmp.Close()
		return fmt.Errorf("seek block: %w", err)
	}
	if _, err := io.CopyN(tmp, src, block.Size); err != nil {
		tmp.Close()
		return fmt.Errorf("copy block: %w", err)
	}

	// Flush & close before rename
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, dst)
}

// Read retrieves a block from storage.
func Read(hash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(config.ObjectsDir, hash+".bin"))
}
