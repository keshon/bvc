package block

import (
	"app/internal/config"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeebo/xxh3"
)

// VerifyBlock checks a single block for integrity using the selected hash.
func (bm *BlockManager) VerifyBlock(hash string) (BlockStatus, error) {
	path := filepath.Join(bm.Root, hash+".bin")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Missing, nil
		}
		return Damaged, err
	}

	var actual string
	switch config.GetSelectedHashName() {
	case "xxh3":
		actual = fmt.Sprintf("%x", xxh3.Hash128(data).Bytes())
	case "sha256":
		h := sha256.Sum256(data)
		actual = fmt.Sprintf("%x", h[:])
	default:
		h := xxh3.Hash128(data).Bytes()
		actual = fmt.Sprintf("%x", h)
	}

	if actual == hash {
		return OK, nil
	}
	return Damaged, nil
}

// Verify checks a set of block hashes concurrently.
func (bm *BlockManager) Verify(hashes map[string]struct{}, workers int) <-chan BlockCheck {
	out := make(chan BlockCheck, 128)
	go func() {
		defer close(out)
		if workers <= 0 {
			workers = 4
		}
		tasks := make(chan string, len(hashes))
		for h := range hashes {
			tasks <- h
		}
		close(tasks)

		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for h := range tasks {
					status, _ := bm.VerifyBlock(h)
					out <- BlockCheck{Hash: h, Status: status}
				}
			}()
		}
		wg.Wait()
	}()
	return out
}
