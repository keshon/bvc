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

	// Normalize slashes
	cleanRel := filepath.ToSlash(filepath.Clean(path))

	// If the input is absolute, convert it to relative before continuing
	var relPath string
	if filepath.IsAbs(cleanRel) {
		rp, err := filepath.Rel(fc.WorkingTreeDir, cleanRel)
		if err != nil {
			return Entry{}, fmt.Errorf("resolve relative path: %w", err)
		}
		relPath = filepath.ToSlash(filepath.Clean(rp))
	} else {
		// Input is already repo-relative (as now returned by Scanner)
		relPath = cleanRel
	}

	// Absolute path that actually exists on disk
	absPath := filepath.Join(fc.WorkingTreeDir, relPath)

	// Split file blocks
	blocks, err := fc.BlockCtx.SplitFile(absPath)
	if err != nil {
		return Entry{}, fmt.Errorf("split %q: %w", relPath, err)
	}

	// Entry.Path MUST stay repo-relative
	return Entry{
		Path:   relPath,
		Blocks: blocks,
	}, nil
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
