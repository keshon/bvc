package file

import (
	"app/internal/progress"
	"app/internal/storage/block"
	"app/internal/util"
	"sync"
)

// Entry represents a file broken into content-defined chunks.
type Entry struct {
	Path   string           `json:"path"`
	Blocks []block.BlockRef `json:"blocks"`
}

// Equal checks if two entries have identical block structures.
func (f *Entry) Equal(other *Entry) bool {
	if f == nil && other == nil {
		return true
	}
	if f == nil || other == nil {
		return false
	}
	if len(f.Blocks) != len(other.Blocks) {
		return false
	}
	for i := range f.Blocks {
		if f.Blocks[i].Hash != other.Blocks[i].Hash ||
			f.Blocks[i].Size != other.Blocks[i].Size {
			return false
		}
	}
	return true
}

// Build splits a single file into content-defined blocks.
func Build(path string) (Entry, error) {
	blocks, err := block.SplitFileIntoBlocks(path)
	if err != nil {
		return Entry{}, err
	}
	return Entry{Path: path, Blocks: blocks}, nil
}

// Store writes all blocks of this file to the object store.
func (e *Entry) Store() error {
	return block.Store(e.Path, e.Blocks)
}

// BuildAll concurrently builds file entries for multiple paths.
func BuildAll(paths []string) ([]Entry, error) {
	bar := progress.NewProgress(len(paths), "Scanning files")
	defer bar.Finish()

	type result struct {
		entry Entry
		err   error
	}

	jobs := make(chan string, len(paths))
	results := make(chan result, len(paths))
	workers := util.WorkerCount()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				entry, err := Build(path)
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

	var entries []Entry
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		entries = append(entries, r.entry)
	}
	return entries, nil
}
