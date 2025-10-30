package block

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"app/internal/config"
	"app/internal/util"

	"golang.org/x/exp/mmap"
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

// SplitFileIntoBlocks splits a file into content-defined blocks
// using dynamic worker count and chunked memory mapping.
func SplitFileIntoBlocks(path string) ([]BlockRef, error) {
	const chunkSize = 1 << 30 // 1 GiB per memory-mapped chunk

	// get file size
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file %q: %w", path, err)
	}
	fileSize := fi.Size()

	var allBlocks []BlockRef
	var offset int64

	// process each chunk sequentially
	for chunkStart := int64(0); chunkStart < fileSize; chunkStart += chunkSize {
		chunkEnd := chunkStart + chunkSize
		if chunkEnd > fileSize {
			chunkEnd = fileSize
		}
		size := int(chunkEnd - chunkStart)

		// mmap this chunk
		reader, err := mmap.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open file %q: %w", path, err)
		}

		data := make([]byte, size)
		if _, err := reader.ReadAt(data, chunkStart); err != nil {
			reader.Close()
			return nil, fmt.Errorf("read mmap file chunk %d-%d: %w", chunkStart, chunkEnd, err)
		}
		reader.Close()

		// CDC logic
		var rh uint32
		start := 0
		var blocks []BlockRef
		var nextIndex int32

		hashCh := make(chan struct {
			data   []byte
			offset int64
		}, 128)

		workers := util.WorkerCount()
		var wg sync.WaitGroup
		wg.Add(workers)
		for w := 0; w < workers; w++ {
			go func() {
				defer wg.Done()
				for item := range hashCh {
					blockRef := hashBlock(item.data, item.offset)
					idx := atomic.AddInt32(&nextIndex, 1) - 1

					if int(idx) >= len(blocks) {
						newBlocks := make([]BlockRef, idx+1)
						copy(newBlocks, blocks)
						blocks = newBlocks
					}
					blocks[idx] = blockRef
				}
			}()
		}

		for i := 0; i < size; i++ {
			rh = (rh<<1 + uint32(data[i])) & 0xFFFFFFFF
			if shouldSplitBlock(i-start+1, rh) {
				hashCh <- struct {
					data   []byte
					offset int64
				}{data[start : i+1], offset}
				offset += int64(i - start + 1)
				start = i + 1
				rh = 0
			}
		}

		if start < size {
			hashCh <- struct {
				data   []byte
				offset int64
			}{data[start:size], offset}
			offset += int64(size - start)
		}

		close(hashCh)
		wg.Wait()

		allBlocks = append(allBlocks, blocks[:nextIndex]...)
	}

	return allBlocks, nil
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
