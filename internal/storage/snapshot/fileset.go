package snapshot

import (
	"fmt"

	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/util"
)

// Fileset represents a snapshot of many files.
type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}

// Build scans all files and constructs a snapshot.
func Build() (Fileset, error) {
	paths, err := file.ListAll()
	if err != nil {
		return Fileset{}, err
	}

	entries, err := file.BuildAll(paths)
	if err != nil {
		return Fileset{}, err
	}

	return Fileset{
		ID:    Hash(entries),
		Files: entries,
	}, nil
}

// Store persists all file blocks belonging to this snapshot.
func (fs *Fileset) Store() error {
	if err := block.CleanupTmp(); err != nil {
		fmt.Printf("Warning: cleanup failed: %v\n", err)
	}
	return util.Parallel(fs.Files, util.WorkerCount(), func(e file.Entry) error {
		return e.Store()
	})
}
