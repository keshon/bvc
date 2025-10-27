package file

import (
	"app/internal/progress"
	"app/internal/storage/block"
	"app/internal/util"
	"os"
	"path/filepath"
	"sync"
)

type Entry struct {
	Path   string
	Blocks []block.BlockRef
}

func (e *Entry) Equal(other *Entry) bool {
	if e == nil && other == nil {
		return true
	}
	if e == nil || other == nil {
		return false
	}
	if len(e.Blocks) != len(other.Blocks) {
		return false
	}
	for i := range e.Blocks {
		if e.Blocks[i].Hash != other.Blocks[i].Hash || e.Blocks[i].Size != other.Blocks[i].Size {
			return false
		}
	}
	return true
}

func Build(path string) (Entry, error) {
	blocks, err := block.SplitFileIntoBlocks(path)
	if err != nil {
		return Entry{}, err
	}
	return Entry{Path: path, Blocks: blocks}, nil
}

func (e *Entry) Store() error {
	return block.Store(e.Path, e.Blocks)
}

func BuildAll(paths []string) ([]Entry, error) {
	bar := progress.NewProgress(len(paths), "Scanning files ")
	defer bar.Finish()

	jobs := make(chan string, len(paths))
	results := make(chan Entry, len(paths))
	errs := make(chan error, len(paths))
	workers := util.WorkerCount()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				entry, err := Build(p)
				if err != nil {
					errs <- err
					continue
				}
				results <- entry
				bar.Increment()
			}
		}()
	}

	for _, p := range paths {
		jobs <- p
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	var entries []Entry
	for entry := range results {
		entries = append(entries, entry)
	}
	if len(errs) > 0 {
		return entries, <-errs
	}
	return entries, nil
}

// BuildTrackedAndUntracked builds entries for all tracked + untracked files.
func BuildTrackedAndUntracked(paths []string) ([]Entry, error) {
	allFiles, err := ListAll()
	if err != nil {
		return nil, err
	}

	// Build entries for everything that currently exists in working directory
	entries, err := BuildAll(allFiles)
	if err != nil {
		return nil, err
	}

	// Also handle deleted files (tracked but no longer exist)
	tracked, _ := GetIndexFiles()
	var deleted []Entry
	for _, t := range tracked {
		if !fileExists(t.Path) {
			deleted = append(deleted, Entry{Path: t.Path, Blocks: nil})
		}
	}

	return append(entries, deleted...), nil
}

// BuildModifiedAndDeleted builds entries only for modified and deleted files.
func BuildModifiedAndDeleted(paths []string) ([]Entry, error) {
	tracked, err := GetIndexFiles()
	if err != nil {
		return nil, err
	}

	var toUpdate []string
	var deleted []Entry

	for _, t := range tracked {
		if !fileExists(t.Path) {
			// tracked file deleted in working tree
			deleted = append(deleted, Entry{Path: t.Path, Blocks: nil})
			continue
		}

		// Compare block hashes to detect modification
		current, err := Build(t.Path)
		if err != nil {
			return nil, err
		}
		if !t.Equal(&current) {
			toUpdate = append(toUpdate, t.Path)
		}
	}

	modified, err := BuildAll(toUpdate)
	if err != nil {
		return nil, err
	}

	return append(modified, deleted...), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(filepath.Clean(path))
	return err == nil
}
