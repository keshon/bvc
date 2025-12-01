package snapshot

import (
	"encoding/json"
	"fmt"

	"path/filepath"
	"sort"

	"github.com/keshon/bvc/internal/progress"

	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/block"
	"github.com/keshon/bvc/internal/repo/file"
	"github.com/keshon/bvc/internal/repo/scan"
	"github.com/keshon/bvc/internal/util"
)

// SnapshotContext handles higher-level operations (filesets, commits)
type SnapshotContext struct {
	SnapshotDir string
	FileCtx     *file.FileContext
	BlockCtx    *block.BlockContext
	Scanner     *scan.Scanner
	FS          fs.FS
}

// Fileset represents a snapshot of tracked files and their block mappings.
type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}

// NewSnapshotContext returns a new SnapshotContext.
func NewSnapshotContext(root string, files *file.FileContext, blocks *block.BlockContext, scanner *scan.Scanner, fs fs.FS) *SnapshotContext {

	return &SnapshotContext{
		SnapshotDir: root,
		FileCtx:     files,
		BlockCtx:    blocks,
		Scanner:     scanner,
		FS:          fs,
	}
}

// BuildWorkingTreeFileset builds a Fileset of tracked (working tree) files.
func (sc *SnapshotContext) BuildTrackedFileset() (Fileset, error) {
	scanResult, err := sc.Scanner.ScanAll()
	if err != nil {
		return Fileset{}, fmt.Errorf("scan tracked files: %w", err)
	}
	return sc.buildFilesetFromPaths(scanResult.Tracked, "tracked")
}

// BuildUntrackedFileset builds a Fileset of untracked files.
func (sc *SnapshotContext) BuildUntrackedFileset() (Fileset, error) {
	scanResult, err := sc.Scanner.ScanAll()
	if err != nil {
		return Fileset{}, fmt.Errorf("scan untracked files: %w", err)
	}
	return sc.buildFilesetFromPaths(scanResult.Untracked, "untracked")
}

// BuildStagedFileset builds a Fileset of staged files.
func (sc *SnapshotContext) BuildStagedFileset() (Fileset, error) {
	scanResult, err := sc.Scanner.ScanAll()
	if err != nil {
		return Fileset{}, fmt.Errorf("scan staged files: %w", err)
	}
	return sc.buildFilesetFromPaths(scanResult.Staged, "staged")
}

// BuildIgnoredFileset builds a Fileset of ignored files.
func (sc *SnapshotContext) BuildIgnoredFileset() (Fileset, error) {
	scanResult, err := sc.Scanner.ScanAll()
	if err != nil {
		return Fileset{}, fmt.Errorf("scan ignored files: %w", err)
	}
	return sc.buildFilesetFromPaths(scanResult.Ignored, "ignored")
}

// BuildAllRepositoryFilesets builds working, staged and ignored filesets in parallel
// and returns them.
func (sc *SnapshotContext) BuildAllRepositoryFilesets() (tracked Fileset, untracked Fileset, staged Fileset, ignored Fileset, err error) {
	type task struct {
		id  int
		run func() (Fileset, error)
	}

	results := make([]Fileset, 4)
	tasks := []task{
		{0, sc.BuildTrackedFileset},
		{1, sc.BuildUntrackedFileset},
		{2, sc.BuildStagedFileset},
		{3, sc.BuildIgnoredFileset},
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
		return Fileset{}, Fileset{}, Fileset{}, Fileset{}, err
	}

	return results[0], results[1], results[2], results[3], nil
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
	return sc.writeJSON(path, fs)
}

// Load retrieves a Fileset by its ID from disk.
func (sc *SnapshotContext) Load(filesetID string) (Fileset, error) {
	path := filepath.Join(sc.SnapshotDir, filesetID+".json")
	var fs Fileset
	if err := sc.readJSON(path, &fs); err != nil {
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
		if err := sc.readJSON(f, &fs); err != nil {
			return nil, fmt.Errorf("failed to read fileset %q: %w", f, err)
		}
		filesets = append(filesets, fs)
	}
	return filesets, nil
}

func (sc *SnapshotContext) writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)

	tmpFile, tmpPath, err := sc.FS.CreateTempFile(dir, "tmp-*.json")
	if err != nil {
		return err
	}
	defer sc.FS.Remove(tmpPath) // ensure cleanup on error

	// write JSON
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// atomically rename
	return sc.FS.Rename(tmpPath, path)
}

func (sc *SnapshotContext) readJSON(path string, v interface{}) error {
	data, err := sc.FS.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
