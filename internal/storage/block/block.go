package block

import (
	"app/internal/config"
	"app/internal/util"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zeebo/xxh3"
)

const (
	minChunkSize = 2 * 1024 * 1024 // 2 MiB
	maxChunkSize = 8 * 1024 * 1024 // 8 MiB
	rollMod      = 4096
)

// SplitFileIntoBlocks divides a file into content-defined chunks.
func SplitFileIntoBlocks(srcPath string) ([]BlockRef, error) {
	f, err := os.Open(srcPath)
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
					blocks = append(blocks, hashBlock(chunk, offset))
					offset += int64(len(chunk))
					chunk = chunk[:0]
				}
			}
		}
		if err != nil {
			if err == os.ErrClosed || err.Error() == "EOF" {
				break
			}
			if err.Error() == "EOF" {
				break
			}
			if err != nil {
				break
			}
		}
		if n == 0 {
			break
		}
	}

	if len(chunk) > 0 {
		blocks = append(blocks, hashBlock(chunk, offset))
	}
	return blocks, nil
}

// hashBlock computes a hash and creates a BlockBlockRef.
func hashBlock(data []byte, offset int64) BlockRef {
	hash := xxh3.Hash128(data).Bytes()
	return BlockRef{
		Hash:   fmt.Sprintf("%x", hash),
		Size:   int64(len(data)),
		Offset: offset,
	}
}

// shouldSplit decides when to end a chunk.
func shouldSplit(size int, rh uint32) bool {
	return (size >= minChunkSize && rh%rollMod == 0) || size >= maxChunkSize
}

func Store(srcPath string, blocks []BlockRef) error {
	return util.Parallel(blocks, util.WorkerCount(), func(b BlockRef) error {
		return writeAtomic(srcPath, b)
	})
}

func writeAtomic(srcPath string, block BlockRef) error {
	dst := filepath.Join(config.ObjectsDir, block.Hash+".bin")
	if fi, err := os.Stat(dst); err == nil && fi.Size() == block.Size {
		return nil
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := make([]byte, block.Size)
	if _, err := f.ReadAt(data, block.Offset); err != nil {
		return fmt.Errorf("read block: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), "tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	if err := tmp.Sync(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	tmp.Close()

	return os.Rename(tmp.Name(), dst)
}

func Read(hash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(config.ObjectsDir, hash+".bin"))
}
