package snapshot

import (
	"fmt"
	"path/filepath"

	"app/internal/progress"
	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/util"
)

// SnapshotManager handles higher-level operations (filesets, commits)
type SnapshotManager struct {
	Root   string
	Files  *file.FileManager
	Blocks *block.BlockManager
}

// Fileset represents a snapshot of tracked files and their block mappings.
type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}

// CreateCurrent builds a Fileset from the current working tree.
func (sm *SnapshotManager) CreateCurrent() (Fileset, error) {
	// Use FileManager to list files and create entries (manager-level operations)
	paths, err := sm.Files.ListAll()
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to list files: %w", err)
	}
	entries, err := sm.Files.CreateEntries(paths)
	if err != nil {
		return Fileset{}, fmt.Errorf("failed to create entries: %w", err)
	}
	return Fileset{
		ID:    HashFileset(entries),
		Files: entries,
	}, nil
}

// Create builds a Fileset from a list of staged entries and stores their blocks.
func (sm *SnapshotManager) Create(entries []file.Entry) (Fileset, error) {
	if len(entries) == 0 {
		return Fileset{}, fmt.Errorf("no files to commit")
	}

	// Store all blocks for the staged files using the BlockManager via the Files layer
	for _, e := range entries {
		// Ensure blocks are written via the FileManager (which uses BlockManager)
		if err := sm.Files.Write(e); err != nil {
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
func (sm *SnapshotManager) Save(fs Fileset) error {
	if fs.ID == "" {
		return fmt.Errorf("invalid fileset: missing ID")
	}
	path := filepath.Join(sm.Root, fs.ID+".json")
	return util.WriteJSON(path, fs)
}

// Load retrieves a Fileset by ID from disk.
func (sm *SnapshotManager) Load(id string) (Fileset, error) {
	path := filepath.Join(sm.Root, id+".json")
	var fs Fileset
	if err := util.ReadJSON(path, &fs); err != nil {
		return Fileset{}, fmt.Errorf("failed to read fileset %q: %w", id, err)
	}
	return fs, nil
}

// List retrieves all filesets from disk.
func (sm *SnapshotManager) List() ([]Fileset, error) {
	files, err := filepath.Glob(filepath.Join(sm.Root, "*.json"))
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
func (sm *SnapshotManager) WriteAndSave(fs *Fileset) error {
	if fs.ID == "" {
		return fmt.Errorf("invalid fileset: missing ID")
	}
	if len(fs.Files) == 0 {
		return fmt.Errorf("invalid fileset: no files")
	}
	if err := sm.writeFiles(fs); err != nil {
		return fmt.Errorf("failed to store files: %w", err)
	}
	return sm.Save(*fs)
}

// writeFiles stores each file’s blocks to disk with progress display.
func (sm *SnapshotManager) writeFiles(fs *Fileset) error {
	// Let BlockManager attempt cleanup of temp files if it provides it.
	// If not present, ignore error (non-fatal).
	if sm.Blocks != nil {
		_ = sm.Blocks.CleanupTemp()
	}

	bar := progress.NewProgress(len(fs.Files), "Storing files ")
	defer bar.Finish()

	return util.Parallel(fs.Files, util.WorkerCount(), func(f file.Entry) error {
		if sm.Blocks == nil || sm.Files == nil {
			return fmt.Errorf("storage managers not attached")
		}
		if err := sm.Blocks.Write(f.Path, f.Blocks); err != nil {
			return fmt.Errorf("error storing file %s: %w", f.Path, err)
		}
		bar.Increment()
		return nil
	})
}
