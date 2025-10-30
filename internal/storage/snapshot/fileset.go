package snapshot

import (
	"fmt"
	"path/filepath"

	"app/internal/config"
	"app/internal/progress"
	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/util"
)

type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}

// CreateCurrentFileset builds a fileset from the current working tree
func CreateCurrentFileset() (Fileset, error) {
	paths, err := file.ListAll()
	if err != nil {
		return Fileset{}, err
	}
	entries, err := file.CreateEntries(paths)
	if err != nil {
		return Fileset{}, err
	}
	return Fileset{
		ID:    HashFileset(entries),
		Files: entries,
	}, nil
}

// GetFileset retrieves a fileset by ID from disk
func GetFileset(id string) (Fileset, error) {
	path := filepath.Join(config.FilesetsDir, id+".json")
	var fs Fileset
	if err := util.ReadJSON(path, &fs); err != nil {
		return Fileset{}, err
	}
	return fs, nil
}

// GetFilesets retrieves all filesets from disk
func GetFilesets() ([]Fileset, error) {
	files, err := filepath.Glob(filepath.Join(config.FilesetsDir, "*.json"))
	if err != nil {
		return nil, err
	}
	var filesets []Fileset
	for _, f := range files {
		var fs Fileset
		if err := util.ReadJSON(f, &fs); err != nil {
			return nil, err
		}
		filesets = append(filesets, fs)
	}
	return filesets, nil
}

// CreateFileset builds a fileset from a list of file entries.
func CreateFileset(entries []file.Entry) (Fileset, error) {
	if len(entries) == 0 {
		return Fileset{}, fmt.Errorf("no files to commit")
	}

	// Store all blocks for the staged files
	for _, e := range entries {
		if err := e.WriteToDisk(); err != nil {
			return Fileset{}, fmt.Errorf("storing file %s: %w", e.Path, err)
		}
	}

	// Compute a fileset hash
	fileset := Fileset{
		Files: entries,
		ID:    HashFileset(entries),
	}

	return fileset, nil
}

// WriteAndSaveFileset stores a fileset to disk
func (fs *Fileset) WriteAndSaveFileset() error {
	if fs.ID == "" {
		return fmt.Errorf("invalid fileset: missing ID")
	}
	if len(fs.Files) == 0 {
		return fmt.Errorf("invalid fileset: no files")
	}
	if err := fs.writeFiles(); err != nil {
		return err
	}
	return SaveFileset(*fs)
}

// saveFileset writes the given Fileset as JSON to the filesets directory.
func SaveFileset(fs Fileset) error {
	if fs.ID == "" {
		return fmt.Errorf("invalid fileset: missing ID")
	}
	path := filepath.Join(config.FilesetsDir, fs.ID+".json")
	return util.WriteJSON(path, fs)
}

// writeFiles stores a fileset to disk
func (fs *Fileset) writeFiles() error {
	if err := block.CleanupTmp(); err != nil {
		fmt.Printf("Warning: cleanup failed: %v\n", err)
	}

	bar := progress.NewProgress(len(fs.Files), "Storing files ")
	defer bar.Finish()

	return util.Parallel(fs.Files, util.WorkerCount(), func(f file.Entry) error {
		if err := block.StoreBlocks(f.Path, f.Blocks); err != nil {
			return err
		}
		bar.Increment()
		return nil
	})
}
