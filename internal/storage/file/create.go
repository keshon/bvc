package file

import (
	"app/internal/progress"
	"app/internal/util"
	"sync"
)

// CreateEntries builds entries from a list of paths.
func (fm *FileManager) CreateEntries(paths []string) ([]Entry, error) {
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
				entry, err := fm.CreateEntry(p)
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

// CreateAllEntries builds entries for all tracked + untracked files.
func (fm *FileManager) CreateAllEntries() ([]Entry, error) {
	allFiles, err := fm.ListAll()
	if err != nil {
		return nil, err
	}
	entries, err := fm.CreateEntries(allFiles)
	if err != nil {
		return nil, err
	}

	tracked, _ := fm.GetIndexFiles()
	var deleted []Entry
	for _, t := range tracked {
		if !fm.Exists(t.Path) {
			deleted = append(deleted, Entry{Path: t.Path, Blocks: nil})
		}
	}
	return append(entries, deleted...), nil
}

// CreateChangedEntries builds entries only for modified and deleted files.
func (fm *FileManager) CreateChangedEntries() ([]Entry, error) {
	tracked, err := fm.GetIndexFiles()
	if err != nil {
		return nil, err
	}

	var toUpdate []string
	var deleted []Entry
	for _, t := range tracked {
		if !fm.Exists(t.Path) {
			deleted = append(deleted, Entry{Path: t.Path, Blocks: nil})
			continue
		}
		current, err := fm.CreateEntry(t.Path)
		if err != nil {
			return nil, err
		}
		if !t.Equal(&current) {
			toUpdate = append(toUpdate, t.Path)
		}
	}

	modified, err := fm.CreateEntries(toUpdate)
	if err != nil {
		return nil, err
	}
	return append(modified, deleted...), nil
}
