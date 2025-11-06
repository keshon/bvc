package util

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"app/internal/fsio"
)

// WriteJSON writes a JSON file atomically to prevent corruption.
var WriteJSON = func(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmpFile, err := fsio.CreateTempFile(dir, "tmp-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer fsio.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return fsio.Rename(tmpPath, path)
}

// ReadJSON reads a JSON file and unmarshals it into v
var ReadJSON = func(path string, v any) error {
	data, err := fsio.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// SortedKeys returns the keys of a map sorted alphabetically.
func SortedKeys[M ~map[string]V, V any](m M) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// WorkerCount returns the number of workers for concurrent operations.
func WorkerCount() int {
	workers := runtime.NumCPU()
	// if runtime.GOOS == "windows" && workers > 2 {
	// 	workers = 2 // limit concurrency on Windows
	// }
	return workers
}

// Parallel runs fn concurrently for each item in inputs, limited by workerLimit.
func Parallel[T any](inputs []T, workerLimit int, fn func(T) error) error {
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
