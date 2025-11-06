package block

import (
	"app/internal/config"
	"crypto/sha256"
	"fmt"

	"github.com/zeebo/xxh3"
)

func shouldSplitBlock(size int, rh uint32) bool {
	return (size >= minChunkSize && rh%rollMod == 0) || size >= maxChunkSize
}

// hashBlock computes the hash of data using the selected algorithm from config.
func hashBlock(data []byte, offset int64) BlockRef {
	var hashStr string
	switch config.SelectedHash() {
	case "xxh3":
		hash := xxh3.Hash128(data).Bytes()
		hashStr = fmt.Sprintf("%x", hash)
	case "sha256":
		hash := sha256.Sum256(data)
		hashStr = fmt.Sprintf("%x", hash[:])
	default:
		// fallback
		hash := xxh3.Hash128(data).Bytes()
		hashStr = fmt.Sprintf("%x", hash)
	}
	return BlockRef{
		Hash:   hashStr,
		Size:   int64(len(data)),
		Offset: offset,
	}
}
