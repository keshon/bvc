package file

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/keshon/bvc/internal/progress"
	"github.com/keshon/bvc/internal/util"
)

// BuildEntry splits a file into block references (content-defined).
func (fc *FileContext) BuildEntry(path string) (Entry, error) {
	if fc.BlockCtx == nil {
		return Entry{}, fmt.Errorf("no BlockContext attached")
	}

	// Normalize just the slashes and cleanliness, not absolute pathing
	cleanPath := filepath.ToSlash(filepath.Clean(path))

	// Compute repository-relative path using the same cleaned value
	relPath, err := filepath.Rel(fc.WorkingTreeDir, cleanPath)
	if err != nil {
		return Entry{}, fmt.Errorf("resolve relative path: %w", err)
	}
	relPath = filepath.ToSlash(relPath)

	// Split based on FS path, not OS absolute path
	blocks, err := fc.BlockCtx.SplitFile(cleanPath)
	if err != nil {
		return Entry{}, fmt.Errorf("split %q: %w", relPath, err)
	}

	return Entry{Path: relPath, Blocks: blocks}, nil
}

// BuildEntries builds entries from a list of paths.
func (fc *FileContext) BuildEntries(paths []string, silent bool) ([]Entry, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	var bar *progress.ProgressTracker
	if !silent {
		bar = progress.NewProgress(len(paths), "Building entries ")
		defer bar.Finish()
	}

	var mu sync.Mutex
	entries := make([]Entry, 0, len(paths))

	err := util.Parallel(paths, util.WorkerCount(), func(p string) error {
		entry, err := fc.BuildEntry(p)
		if err != nil {
			return err
		}

		mu.Lock()
		entries = append(entries, entry)
		mu.Unlock()

		if !silent {
			bar.Increment()
		}
		return nil
	})

	if err != nil {
		return entries, err
	}

	return entries, nil
}

// Write stores all blocks of an entry into store.
func (fc *FileContext) Write(e Entry) error {
	if fc.BlockCtx == nil {
		return fmt.Errorf("no BlockContext attached")
	}
	return fc.BlockCtx.Write(e.Path, e.Blocks)
}

// Exists checks whether a given path exists in the working tree.
func (fc *FileContext) Exists(path string) bool {
	_, err := fc.FS.Stat(filepath.Clean(path))
	return err == nil
}
