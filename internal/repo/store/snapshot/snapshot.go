package snapshot

import (
	"fmt"
	"path/filepath"
	"sort"

	"app/internal/fsio"
	"app/internal/progress"
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"

	"app/internal/util"
)

// SnapshotContext handles higher-level operations (filesets, commits)
type SnapshotContext struct {
	Root   string
	Files  *file.FileContext
	Blocks *block.BlockContext
}

// Fileset represents a snapshot of tracked files and their block mappings.
type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}

// CreateCurrent builds a Fileset from the current working tree.
func (sc *SnapshotContext) CreateCurrent() (Fileset, error) {
	paths, err := sc.Files.ListAll()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list files: %w", err)
	}

	entries, err := sc.Files.CreateEntries(paths)
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to create entries: %w", err)
	}

	// Sort entries by path before computing hash
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })

	return Fileset{
		ID:    HashFileset(entries),
		Files: entries,
	}, nil
}

// Create builds a Fileset from a list of staged entries and stores their blocks.
func (sc *SnapshotContext) Create(entries []file.Entry) (Fileset, error) {
	if len(entries) == 0 {
		return Fileset{}, fmt.Errorf("no files to commit")
	}

	// Store all blocks for the staged files using the BlockManager via the Files layer
	for _, e := range entries {
		// Ensure blocks are written via the FileManager (which uses BlockManager)
		if err := sc.Files.Write(e); err != nil {
			return Fileset{}, fmt.Errorf("storing file %s: %w", e.Path, err)
		}
	}

	fs := Fileset{
		ID:    HashFileset(entries),
		Files: entries,
	}
	return fs, nil
}

// Save persists a Fileset JSON to disk.
func (sc *SnapshotContext) Save(fs Fileset) error {
	if fs.ID == "" {
		return fmt.Errorf("invalid fileset: missing ID")
	}

	if err := fsio.MkdirAll(sc.Root, 0o755); err != nil {
		return fmt.Errorf("create snapshots dir: %w", err)
	}

	path := filepath.Join(sc.Root, fs.ID+".json")
	return util.WriteJSON(path, fs)
}

// Load retrieves a Fileset by its ID from disk.
func (sc *SnapshotContext) Load(filesetID string) (Fileset, error) {
	path := filepath.Join(sc.Root, filesetID+".json")
	var fs Fileset
	if err := util.ReadJSON(path, &fs); err != nil {
		return Fileset{}, fmt.Errorf("failed to read fileset %q: %w", filesetID, err)
	}
	return fs, nil
}

// List retrieves all filesets from disk.
func (sc *SnapshotContext) List() ([]Fileset, error) {
	files, err := filepath.Glob(filepath.Join(sc.Root, "*.json"))
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
