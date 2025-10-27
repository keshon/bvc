package block

import (
	"app/internal/config"
	"app/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeebo/xxh3"
)

func Verify(hash string) (BlockStatus, error) {
	path := filepath.Join(config.ObjectsDir, hash+".bin")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Missing, nil
		}
		return Missing, err
	}

	if fmt.Sprintf("%x", xxh3.Hash128(data).Bytes()) == hash {
		return OK, nil
	}
	return Damaged, nil
}

func VerifyMany(blocks map[string]struct{}, workers int) <-chan BlockCheck {
	out := make(chan BlockCheck, 128)

	go func() {
		defer close(out)
		tasks := make(chan string, len(blocks))
		for h := range blocks {
			tasks <- h
		}
		close(tasks)

		var wg sync.WaitGroup
		if workers <= 0 {
			workers = util.WorkerCount()
		}

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for h := range tasks {
					BlockStatus, _ := Verify(h)
					out <- BlockCheck{Hash: h, BlockStatus: BlockStatus}
				}
			}()
		}

		wg.Wait()
	}()
	return out
}
