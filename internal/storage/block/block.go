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
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}
	defer f.Close()

	var (
		blocks []BlockRef
		buf    = make([]byte, maxChunkSize)
		chunk  = make([]byte, 0, maxChunkSize) // pre-allocate
		offset int64
		rh     uint32
	)

	for {
		n, err := f.Read(buf)
		if n > 0 {
			start := 0
			for i := 0; i < n; i++ {
				rh = (rh<<1 + uint32(buf[i])) & 0xFFFFFFFF
				if shouldSplitBlock(i-start+1+len(chunk), rh) {
					// append previous chunk + current slice
					blockData := make([]byte, len(chunk)+(i-start+1))
					copy(blockData, chunk)
					copy(blockData[len(chunk):], buf[start:i+1])

					blocks = append(blocks, hashBlock(blockData, offset))
					offset += int64(len(blockData))

					// reset for next block
					chunk = chunk[:0]
					start = i + 1
					rh = 0
				}
			}
			// append leftover bytes to chunk for next read
			if start < n {
				chunk = append(chunk, buf[start:n]...)
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read file %q: %w", path, err)
		}

		if n == 0 {
			break
		}
	}

	// handle remaining data
	if len(chunk) > 0 {
		blocks = append(blocks, hashBlock(chunk, offset))
	}

	return blocks, nil
}

// StoreBlocks writes all blocks concurrently and safely.
func StoreBlocks(filePath string, blocks []BlockRef) error {
	workers := util.WorkerCount()
	return util.Parallel(blocks, workers, func(b BlockRef) error {
		return writeBlockAtomic(filePath, b)
	})
}

// writeBlockAtomic writes a single block atomically to storage.
func writeBlockAtomic(filePath string, block BlockRef) error {
	dst := filepath.Join(config.ObjectsDir, block.Hash+".bin")

	// Fast path — block already exists
	if fi, err := os.Stat(dst); err == nil && fi.Size() == block.Size {
		return nil
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ensure dir for %q: %w", dst, err)
	}

	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", filePath, err)
	}
	defer src.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %q: %w", filepath.Dir(dst), err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	// Stream copy the block
	if _, err := src.Seek(block.Offset, io.SeekStart); err != nil {
		tmp.Close()
		return fmt.Errorf("seek to offset %d in %q: %w", block.Offset, filePath, err)
	}
	if _, err := io.CopyN(tmp, src, block.Size); err != nil && !errors.Is(err, io.EOF) {
		tmp.Close()
		return fmt.Errorf("copy block %q: %w", block.Hash, err)
	}

	// Flush & close before rename
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp file %q: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file %q: %w", tmpPath, err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("rename temp file %q to %q: %w", tmpPath, dst, err)
	}

	// TODO: optional: verify block integrity here

	return nil
}

// ReadBlock retrieves a block from storage.
func ReadBlock(hash string) ([]byte, error) {
	path := filepath.Join(config.ObjectsDir, hash+".bin")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read block %q: %w", hash, err)
	}
	return data, nil
}
