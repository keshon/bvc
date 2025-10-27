package snapshot

import (
	"fmt"

	"app/internal/progress"
	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/util"
)

type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}

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

// func (fs *Fileset) Store() error {
// 	if err := block.CleanupTmp(); err != nil {
// 		fmt.Printf("Warning: cleanup failed: %v\n", err)
// 	}
// 	return util.Parallel(fs.Files, util.WorkerCount(), func(e file.Entry) error {
// 		return e.Store()
// 	})
// }

func (fs *Fileset) Store() error {
	if err := block.CleanupTmp(); err != nil {
		fmt.Printf("Warning: cleanup failed: %v\n", err)
	}

	// Count total blocks for the progress bar
	totalBlocks := 0
	for _, f := range fs.Files {
		totalBlocks += len(f.Blocks)
	}

	bar := progress.NewProgress(totalBlocks, "Storing blocks")
	defer bar.Finish()

	return util.Parallel(fs.Files, util.WorkerCount(), func(f file.Entry) error {
		for _, b := range f.Blocks {
			if err := block.WriteAtomic(f.Path, b); err != nil {
				return err
			}
			bar.Increment()
		}
		return nil
	})
}
