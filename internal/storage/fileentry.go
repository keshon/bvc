package storage

import (
	"app/internal/progress"
	"sync"
)

// BuildFileEntry splits a file into content-defined blocks.
func BuildFileEntry(path string) (FileEntry, error) {
	blocks, err := SplitFileIntoBlocks(path)
	if err != nil {
		return FileEntry{}, err
	}
	return FileEntry{Path: path, Blocks: blocks}, nil
}

// StoreFileEntry writes all its blocks concurrently.
func StoreFileEntry(entry FileEntry) error {
	return StoreBlocks(entry.Path, entry.Blocks)
}

// buildFileEntries concurrently builds file entries for multiple paths.
func buildFileEntries(paths []string) ([]FileEntry, error) {
	bar := progress.NewProgress(len(paths), "Scanning FileEntries")
	defer bar.Finish()

	type result struct {
		entry FileEntry
		err   error
	}

	jobs := make(chan string, len(paths))
	results := make(chan result, len(paths))
	workers := WorkerCount()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				entry, err := BuildFileEntry(path)
				results <- result{entry, err}
				bar.Increment()
			}
		}()
	}

	for _, path := range paths {
		jobs <- path
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var files []FileEntry
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		files = append(files, r.entry)
	}
	return files, nil
}
