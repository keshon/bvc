package storage

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/zeebo/xxh3"
)

// workerCount returns the number of workers for concurrent operations.
func WorkerCount() int {
	return min(runtime.NumCPU(), 4)
}

// ComputeFilesetID generates a deterministic ID from file entries.
func HashFileset(files []FileEntry) string {
	// deterministic order
	paths := make([]string, 0, len(files))
	m := make(map[string]FileEntry, len(files))
	for _, f := range files {
		p := filepath.Clean(f.Path)
		paths = append(paths, p)
		m[p] = f
	}
	sort.Strings(paths)

	data := []byte{}
	for _, p := range paths {
		for _, b := range m[p].Blocks {
			data = append(data, []byte(b.Hash)...)
		}
	}

	return fmt.Sprintf("%x", xxh3.Hash128(data).Bytes())
}

// parallel runs fn concurrently for each item in inputs, limited by workerLimit.
func parallel[T any](inputs []T, workerLimit int, fn func(T) error) error {
	if len(inputs) == 0 {
		return nil
	}

	sem := make(chan struct{}, workerLimit)
	errCh := make(chan error, len(inputs))
	var wg sync.WaitGroup

	for _, in := range inputs {
		sem <- struct{}{}
		wg.Add(1)
		go func(x T) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := fn(x); err != nil {
				errCh <- err
			}
		}(in)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}
	return nil
}
