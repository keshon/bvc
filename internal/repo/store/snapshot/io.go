package snapshot

import (
	"fmt"
	"path/filepath"

	"app/internal/fsio"
	"app/internal/util"
)

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
