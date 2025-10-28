package block

import (
	"fmt"

	"github.com/zeebo/xxh3"
)

func shouldSplitBlock(size int, rh uint32) bool {
	return (size >= minChunkSize && rh%rollMod == 0) || size >= maxChunkSize
}

func hashBlock(data []byte, offset int64) BlockRef {
	hash := xxh3.Hash128(data).Bytes()
	return BlockRef{
		Hash:   fmt.Sprintf("%x", hash),
		Size:   int64(len(data)),
		Offset: offset,
	}
}
