package block

import (
	"fmt"

	"github.com/zeebo/xxh3"
)

func ShouldSplitBlock(size int, rh uint32) bool {
	return (size >= minChunkSize && rh%rollMod == 0) || size >= maxChunkSize
}

// hashBlock computes the hash of data using the selected algorithm from config.
func HashBlock(data []byte, offset int64) BlockRef {
	var hashStr string

	hash := xxh3.Hash128(data).Bytes()
	hashStr = fmt.Sprintf("%x", hash)

	return BlockRef{
		Hash:   hashStr,
		Size:   int64(len(data)),
		Offset: offset,
	}
}
