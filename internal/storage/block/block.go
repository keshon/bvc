package block

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"app/internal/config"
	"app/internal/util"
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
				if shouldSplitBlock(len(chunk), rh) {
					blocks = append(blocks, hashBlock(chunk, offset))
					offset += int64(len(chunk))
					chunk = chunk[:0]
				}
			}
		}
		if err != nil {
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, err
			}
			return nil, err
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

// Store writes all blocks concurrently and safely on Windows.
func StoreBlocks(filePath string, blocks []BlockRef) error {
	workers := util.WorkerCount()
	return util.Parallel(blocks, workers, func(b BlockRef) error {
		return writeBlockAtomic(filePath, b)
	})
}

// WriteAtomic writes a single block to the object store atomically.
func writeBlockAtomic(filePath string, block BlockRef) error {
	dst := filepath.Join(config.ObjectsDir, block.Hash+".bin")

	// Fast path — block already exists
	if fi, err := os.Stat(dst); err == nil && fi.Size() == block.Size {
		return nil
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ensure dir: %w", err)
	}

	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	// Stream copy the block
	if _, err := src.Seek(block.Offset, io.SeekStart); err != nil {
		tmp.Close()
		return fmt.Errorf("seek block: %w", err)
	}
	if _, err := io.CopyN(tmp, src, block.Size); err != nil && err != io.EOF {
		tmp.Close()
		return fmt.Errorf("copy block: %w", err)
	}

	// Flush & close before rename
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	// Atomic rename to final destination
	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("rename temp: %w", err)
	}

	// Verify block integrity using existing Verify()
	// status, err := Verify(block.Hash)
	// if err != nil {
	// 	return fmt.Errorf("verify block: %w", err)
	// }
	// if status != OK {
	// 	return fmt.Errorf("verify block: integrity check failed for %s", block.Hash)
	// }

	return nil
}

// Read retrieves a block from storage.
func ReadBlock(hash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(config.ObjectsDir, hash+".bin"))
}
