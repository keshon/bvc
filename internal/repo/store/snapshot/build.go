package snapshot

import (
	"app/internal/progress"
	"app/internal/repo/store/file"
	"app/internal/util"
	"fmt"
	"sort"
)

// TODO: split into smaller functions and make one unified
// BuildFilesetFromWorkingTree builds a Fileset from the current working tree.
func (sc *SnapshotContext) BuildFilesetsFromWorkingTree() (tracked Fileset, staged Fileset, ignored Fileset, err error) {
	trackedPaths, stagedPaths, ignoredPaths, err := sc.Files.ScanFilesInWorkingTree()
	if err != nil {
		return Fileset{}, Fileset{}, Fileset{}, fmt.Errorf("failed to list files: %w", err)
	}

	trackedEntries, err := sc.Files.BuildEntries(trackedPaths, true)
	if err != nil {
		return Fileset{}, Fileset{}, Fileset{}, fmt.Errorf("failed to create tracked entries: %w", err)
	}

	stagedEntries, err := sc.Files.BuildEntries(stagedPaths, true)
	if err != nil {
		return Fileset{}, Fileset{}, Fileset{}, fmt.Errorf("failed to create staged entries: %w", err)
	}

	ignoredEntries, err := sc.Files.BuildEntries(ignoredPaths, true)
	if err != nil {
		return Fileset{}, Fileset{}, Fileset{}, fmt.Errorf("failed to create ignored entries: %w", err)
	}

	sort.Slice(trackedEntries, func(i, j int) bool { return trackedEntries[i].Path < trackedEntries[j].Path })
	sort.Slice(stagedEntries, func(i, j int) bool { return stagedEntries[i].Path < stagedEntries[j].Path })
	sort.Slice(ignoredEntries, func(i, j int) bool { return ignoredEntries[i].Path < ignoredEntries[j].Path })

	return Fileset{
			ID:    HashFileset(trackedEntries),
			Files: trackedEntries,
		}, Fileset{
			ID:    HashFileset(stagedEntries),
			Files: stagedEntries,
		},
		Fileset{
			ID:    HashFileset(ignoredEntries),
			Files: ignoredEntries,
		}, nil
}

// BuildFilesetFromEntries builds a Fileset from staged entries and stores their blocks.
func (sc *SnapshotContext) BuildFilesetFromEntries(entries []file.Entry) (Fileset, error) {
	if len(entries) == 0 {
		return Fileset{}, fmt.Errorf("no files to commit")
	}

	for _, e := range entries {
		if err := sc.Files.Write(e); err != nil {
			return Fileset{}, fmt.Errorf("storing file %s: %w", e.Path, err)
		}
	}

	return Fileset{ID: HashFileset(entries), Files: entries}, nil
}

// WriteAndSave stores all file blocks and saves the Fileset metadata.
func (sc *SnapshotContext) WriteAndSave(fs *Fileset) error {
	if fs.ID == "" {
		return fmt.Errorf("invalid fileset: missing ID")
	}
	if len(fs.Files) == 0 {
		return fmt.Errorf("invalid fileset: no files")
	}
	if err := sc.writeFiles(fs); err != nil {
		return fmt.Errorf("failed to store files: %w", err)
	}
	return sc.Save(*fs)
}

// writeFiles stores each file’s blocks to disk with progress display.
func (sc *SnapshotContext) writeFiles(fs *Fileset) error {
	// Let BlockManager attempt cleanup of temp files if it provides it.
	// If not present, ignore error (non-fatal).
	if sc.Blocks != nil {
		_ = sc.Blocks.CleanupTemp()
	}

	bar := progress.NewProgress(len(fs.Files), "Storing files ")
	defer bar.Finish()

	return util.Parallel(fs.Files, util.WorkerCount(), func(f file.Entry) error {
		if sc.Blocks == nil || sc.Files == nil {
			return fmt.Errorf("store managers not attached")
		}
		if err := sc.Blocks.Write(f.Path, f.Blocks); err != nil {
			return fmt.Errorf("error storing file %s: %w", f.Path, err)
		}
		bar.Increment()
		return nil
	})
}
