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

// BuildFileset builds a fileset from the current working tree
func BuildFileset() (Fileset, error) {
	paths, err := file.ListAll()
	if err != nil {
		return Fileset{}, err
	}
	entries, err := file.BuildEntries(paths)
	if err != nil {
		return Fileset{}, err
	}
	return Fileset{
		ID:    HashFileset(entries),
		Files: entries,
	}, nil
}

// LoadFileset retrieves a fileset by ID from disk
func LoadFileset(id string) (Fileset, error) {
	path := filepath.Join(config.FilesetsDir, id+".json")
	var fs Fileset
	if err := util.ReadJSON(path, &fs); err != nil {
		return Fileset{}, err
	}
	return fs, nil
}

// BuildFilesetFromEntries builds a fileset from a list of file entries.
func BuildFilesetFromEntries(entries []file.Entry) (Fileset, error) {
	if len(entries) == 0 {
		return Fileset{}, fmt.Errorf("no files to commit")
	}

	// Store all blocks for the staged files
	for _, e := range entries {
		if err := e.Store(); err != nil {
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

func (fs *Fileset) Store() error {
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
