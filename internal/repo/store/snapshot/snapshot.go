package snapshot

import (
	"app/internal/fs"
	"app/internal/progress"
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
	"app/internal/util"
	"fmt"
	"path/filepath"
	"sort"
)

// SnapshotContext handles higher-level operations (filesets, commits)
type SnapshotContext struct {
	SnapshotDir string
	FileCtx     *file.FileContext
	BlockCtx    *block.BlockContext
	FS          fs.FS
}

// Fileset represents a snapshot of tracked files and their block mappings.
type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}

// NewSnapshotContext returns a new SnapshotContext.
func NewSnapshotContext(root string, files *file.FileContext, blocks *block.BlockContext, fs fs.FS) *SnapshotContext {
	return &SnapshotContext{SnapshotDir: root, FileCtx: files, BlockCtx: blocks, FS: fs}
}

// BuildWorkingTreeFileset builds a Fileset of tracked (working tree) files.
func (sc *SnapshotContext) BuildWorkingTreeFileset() (Fileset, error) {
	trackedPaths, _, _, err := sc.FileCtx.ScanAllRepository()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list tracked files: %w", err)
	}
	return sc.buildFilesetFromPaths(trackedPaths, "tracked")
}

// BuildStagedFileset builds a Fileset of staged files.
func (sc *SnapshotContext) BuildStagedFileset() (Fileset, error) {
	_, stagedPaths, _, err := sc.FileCtx.ScanAllRepository()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list staged files: %w", err)
	}
	return sc.buildFilesetFromPaths(stagedPaths, "staged")
}

// BuildIgnoredFileset builds a Fileset of ignored files.
func (sc *SnapshotContext) BuildIgnoredFileset() (Fileset, error) {
	_, _, ignoredPaths, err := sc.FileCtx.ScanAllRepository()
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

	entries, err := sc.FileCtx.BuildEntries(paths, true)
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
		if err := sc.FileCtx.Write(e); err != nil {
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
	if sc.BlockCtx != nil {
		_ = sc.BlockCtx.CleanupTemp()
	}

	bar := progress.NewProgress(len(fs.Files), "Storing files ")
	defer bar.Finish()

	return util.Parallel(fs.Files, util.WorkerCount(), func(f file.Entry) error {
		if sc.BlockCtx == nil || sc.FileCtx == nil {
			return fmt.Errorf("store managers not attached")
		}
		if err := sc.BlockCtx.Write(f.Path, f.Blocks); err != nil {
			return fmt.Errorf("error storing file %s: %w", f.Path, err)
		}
		bar.Increment()
		return nil
	})
}

// Save persists a Fileset JSON to disk.
func (sc *SnapshotContext) Save(fs Fileset) error {
	if fs.ID == "" {
		return fmt.Errorf("invalid fileset: missing ID")
	}

	if err := sc.FS.MkdirAll(sc.SnapshotDir, 0o755); err != nil {
		return fmt.Errorf("create snapshots dir: %w", err)
	}

	path := filepath.Join(sc.SnapshotDir, fs.ID+".json")
	return util.WriteJSON(path, fs)
}

// Load retrieves a Fileset by its ID from disk.
func (sc *SnapshotContext) Load(filesetID string) (Fileset, error) {
	path := filepath.Join(sc.SnapshotDir, filesetID+".json")
	var fs Fileset
	if err := util.ReadJSON(path, &fs); err != nil {
		return Fileset{}, fmt.Errorf("failed to read fileset %q: %w", filesetID, err)
	}
	return fs, nil
}

// List retrieves all filesets from disk.
func (sc *SnapshotContext) List() ([]Fileset, error) {
	files, err := filepath.Glob(filepath.Join(sc.SnapshotDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list filesets: %w", err)
	}
	var filesets []Fileset
	for _, f := range files {
		var fs Fileset
		if err := util.ReadJSON(f, &fs); err != nil {
			return nil, fmt.Errorf("failed to read fileset %q: %w", f, err)
		}
		filesets = append(filesets, fs)
	}
	return filesets, nil
}
