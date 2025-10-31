package block

import (
	"app/internal/util"
	"fmt"
	"os"
	"sync"

	"golang.org/x/exp/mmap"
)

// SplitFile divides a file into content-defined blocks.
func (bm *BlockManager) SplitFile(path string) ([]BlockRef, error) {
	const chunkSize = 1 << 30 // 1 GiB per memory-mapped chunk

	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file %q: %w", path, err)
	}
	fileSize := fi.Size()

	var allBlocks []BlockRef
	var offset int64

	for chunkStart := int64(0); chunkStart < fileSize; chunkStart += chunkSize {
		chunkEnd := chunkStart + chunkSize
		if chunkEnd > fileSize {
			chunkEnd = fileSize
		}
		size := int(chunkEnd - chunkStart)

		reader, err := mmap.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open file %q: %w", path, err)
		}

		data := make([]byte, size)
		if _, err := reader.ReadAt(data, chunkStart); err != nil {
			reader.Close()
			return nil, fmt.Errorf("read mmap chunk %d-%d: %w", chunkStart, chunkEnd, err)
		}
		reader.Close()

		var rh uint32
		start := 0

		hashCh := make(chan struct {
			data   []byte
			offset int64
		}, 128)

		var mu sync.Mutex
		workers := util.WorkerCount()
		var wg sync.WaitGroup
		wg.Add(workers)

		for range workers {
			go func() {
				defer wg.Done()
				for item := range hashCh {
					blockRef := hashBlock(item.data, item.offset)
					mu.Lock()
					allBlocks = append(allBlocks, blockRef)
					mu.Unlock()
				}
			}()
		}

		for i := range size {
			rh = (rh<<1 + uint32(data[i])) & 0xFFFFFFFF
			if shouldSplitBlock(i-start+1, rh) {
				blockSlice := data[start : i+1]
				hashCh <- struct {
					data   []byte
					offset int64
				}{blockSlice, offset}

				offset += int64(i - start + 1)
				start = i + 1
				rh = 0
			}
		}

		if start < size {
			blockSlice := data[start:size]
			hashCh <- struct {
				data   []byte
				offset int64
			}{blockSlice, offset}
			offset += int64(size - start)
		}

		close(hashCh)
		wg.Wait()
	}

	return allBlocks, nil
}
