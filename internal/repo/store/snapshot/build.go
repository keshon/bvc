package snapshot

import (
	"app/internal/progress"
	"app/internal/repo/store/file"
	"app/internal/util"
	"fmt"
	"sort"
)

// BuildFilesetFromWorkingTree builds a Fileset from the current working tree.
func (sc *SnapshotContext) BuildFilesetFromWorkingTree() (Fileset, error) {
	paths, err := sc.Files.ScanFilesInWorkingTree()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list files: %w", err)
	}

	entries, err := sc.Files.BuildEntries(paths)
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to create entries: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return Fileset{ID: HashFileset(entries), Files: entries}, nil
}

// BuildFilesetFromStaged builds a Fileset from staged entries and stores their blocks.
func (sc *SnapshotContext) BuildFilesetFromStaged(entries []file.Entry) (Fileset, error) {
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
