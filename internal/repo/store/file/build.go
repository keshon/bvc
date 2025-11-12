package file

import (
	"app/internal/fsio"
	"app/internal/progress"
	"app/internal/util"
	"fmt"
	"path/filepath"
	"sync"
)

// BuildEntry splits a file into block references (content-defined).
func (fc *FileContext) BuildEntry(path string) (Entry, error) {
	if fc.Blocks == nil {
		return Entry{}, fmt.Errorf("no BlockContext attached")
	}
	blocks, err := fc.Blocks.SplitFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("split %q: %w", path, err)
	}
	return Entry{Path: path, Blocks: blocks}, nil
}

// BuildEntries builds entries from a list of paths.
func (fc *FileContext) BuildEntries(paths []string) ([]Entry, error) {
	bar := progress.NewProgress(len(paths), "Scanning files ")
	defer bar.Finish()

	jobs := make(chan string, len(paths))
	results := make(chan Entry, len(paths))
	errs := make(chan error, len(paths))
	workers := util.WorkerCount()

	var wg sync.WaitGroup
	wg.Add(workers)

	// Start workers
	for range workers {
		go func() {
			defer wg.Done()
			for p := range jobs {
				entry, err := fc.BuildEntry(p)
				if err != nil {
					errs <- err
					continue
				}
				results <- entry
				bar.Increment()
			}
		}()
	}

	// Send jobs
	for _, p := range paths {
		jobs <- p
	}
	close(jobs)

	// Wait for workers to finish, then close channels
	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	var entries []Entry
	for e := range results {
		entries = append(entries, e)
	}

	// Collect one error if any
	select {
	case err := <-errs:
		return entries, err
	default:
	}

	return entries, nil
}

// BuildAllEntries builds entries for all tracked + untracked files.
func (fc *FileContext) BuildAllEntries() ([]Entry, error) {
	paths, _, err := fc.ScanFilesInWorkingTree()
	if err != nil {
		return nil, err
	}
	entries, err := fc.BuildEntries(paths)
	if err != nil {
		return nil, err
	}

	tracked, _ := fc.LoadIndex()
	var deleted []Entry
	for _, t := range tracked {
		if !fc.Exists(t.Path) {
			deleted = append(deleted, Entry{Path: t.Path, Blocks: nil})
		}
	}
	return append(entries, deleted...), nil
}

// BuildChangedEntries builds entries only for modified and deleted files.
func (fc *FileContext) BuildChangedEntries() ([]Entry, error) {
	tracked, err := fc.LoadIndex()
	if err != nil {
		return nil, err
	}

	var toUpdate []string
	var deleted []Entry
	for _, t := range tracked {
		if !fc.Exists(t.Path) {
			deleted = append(deleted, Entry{Path: t.Path, Blocks: nil})
			continue
		}
		current, err := fc.BuildEntry(t.Path)
		if err != nil {
			return nil, err
		}
		if !t.Equal(&current) {
			toUpdate = append(toUpdate, t.Path)
		}
	}

	modified, err := fc.BuildEntries(toUpdate)
	if err != nil {
		return nil, err
	}
	return append(modified, deleted...), nil
}

// Write stores all blocks of an entry into store.
func (fc *FileContext) Write(e Entry) error {
	if fc.Blocks == nil {
		return fmt.Errorf("no BlockContext attached")
	}
	return fc.Blocks.Write(e.Path, e.Blocks)
}

// Exists checks whether a given path exists in the working tree.
func (fc *FileContext) Exists(path string) bool {
	_, err := fsio.StatFile(filepath.Clean(path))
	return err == nil
}
