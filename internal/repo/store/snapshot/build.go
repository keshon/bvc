package snapshot

import (
	"app/internal/progress"
	"app/internal/repo/store/file"
	"app/internal/util"
	"fmt"
	"sort"
)

// BuildWorkingTreeFileset builds a Fileset of tracked (working tree) files.
func (sc *SnapshotContext) BuildWorkingTreeFileset() (Fileset, error) {
	trackedPaths, _, _, err := sc.Files.ScanFilesInWorkingTree()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list tracked files: %w", err)
	}
	return sc.buildFilesetFromPaths(trackedPaths, "tracked")
}

// BuildStagedFileset builds a Fileset of staged files.
func (sc *SnapshotContext) BuildStagedFileset() (Fileset, error) {
	_, stagedPaths, _, err := sc.Files.ScanFilesInWorkingTree()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list staged files: %w", err)
	}
	return sc.buildFilesetFromPaths(stagedPaths, "staged")
}

// BuildIgnoredFileset builds a Fileset of ignored files.
func (sc *SnapshotContext) BuildIgnoredFileset() (Fileset, error) {
	_, _, ignoredPaths, err := sc.Files.ScanFilesInWorkingTree()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list ignored files: %w", err)
	}
	return sc.buildFilesetFromPaths(ignoredPaths, "ignored")
}

// BuildAllRepositoryFilesets builds working, staged and ignored filesets in parallel
// and returns them.
func (sc *SnapshotContext) BuildAllRepositoryFilesets() (tracked Fileset, staged Fileset, ignored Fileset, err error) {
	type task struct {
		id  int
		run func() (Fileset, error)
	}

	results := make([]Fileset, 3)
	tasks := []task{
		{0, sc.BuildWorkingTreeFileset},
		{1, sc.BuildStagedFileset},
		{2, sc.BuildIgnoredFileset},
	}

	err = util.Parallel(tasks, len(tasks), func(t task) error {
		fs, e := t.run()
		if e != nil {
			return e
		}
		results[t.id] = fs
		return nil
	})
	if err != nil {
		return Fileset{}, Fileset{}, Fileset{}, err
	}

	return results[0], results[1], results[2], nil
}

// buildFilesetFromPaths is a small helper to avoid duplication.
func (sc *SnapshotContext) buildFilesetFromPaths(paths []string, label string) (Fileset, error) {
	if len(paths) == 0 {
		return Fileset{Files: nil}, nil
	}

	entries, err := sc.Files.BuildEntries(paths, true)
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to build %s entries: %w", label, err)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })

	return Fileset{
		ID:    HashFileset(entries),
		Files: entries,
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

	return Fileset{
		ID:    HashFileset(entries),
		Files: entries,
	}, nil
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
