package block

import (
	"fmt"
	"os"
)

// SplitFile divides a file into content-defined blocks deterministically.
func (bc *BlockContext) SplitFile(path string) ([]BlockRef, error) {
	const chunkSize = 1 << 30 // 1 GiB per chunk

	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file %q: %w", path, err)
	}
	fileSize := fi.Size()

	var allBlocks []BlockRef
	var offset int64

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}
	defer file.Close()

	for chunkStart := int64(0); chunkStart < fileSize; chunkStart += chunkSize {
		chunkEnd := chunkStart + chunkSize
		if chunkEnd > fileSize {
			chunkEnd = fileSize
		}
		size := int(chunkEnd - chunkStart)

		data := make([]byte, size)
		if _, err := file.ReadAt(data, chunkStart); err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("read file chunk %d-%d: %w", chunkStart, chunkEnd, err)
		}

		var rh uint32
		start := 0

		for i := 0; i < size; i++ {
			rh = (rh<<1 + uint32(data[i])) & 0xFFFFFFFF
			if ShouldSplitBlock(i-start+1, rh) {
				blockSlice := data[start : i+1]
				allBlocks = append(allBlocks, HashBlock(blockSlice, offset))
				offset += int64(i - start + 1)
				start = i + 1
				rh = 0
			}
		}

		// Remaining block at the end of the chunk
		if start < size {
			blockSlice := data[start:size]
			allBlocks = append(allBlocks, HashBlock(blockSlice, offset))
			offset += int64(size - start)
		}
	}

	return allBlocks, nil
}
