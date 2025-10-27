package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

// WriteJSON writes a JSON file atomically to prevent corruption.
func WriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, "tmp-*.json")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

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

	return os.Rename(tmpPath, path)
}

// ReadJSON reads a JSON file and unmarshals it into v
func ReadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
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
	return min(runtime.NumCPU(), 4)
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
