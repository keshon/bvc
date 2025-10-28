package block

import (
	"app/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeebo/xxh3"
)

func VerifyBlock(hash string) (BlockStatus, error) {
	path := filepath.Join(config.ObjectsDir, hash+".bin")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Missing, nil
		}
		return Damaged, err
	}
	if fmt.Sprintf("%x", xxh3.Hash128(data).Bytes()) == hash {
		return OK, nil
	}
	return Damaged, nil
}

func VerifyBlocks(hashes map[string]struct{}, workers int) <-chan BlockCheck {
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
					status, _ := VerifyBlock(h)
					out <- BlockCheck{Hash: h, Status: status}
				}
			}()
		}
		wg.Wait()
	}()
	return out
}
