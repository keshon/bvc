package verify

import (
	"app/internal/config"
	"app/internal/core"
	"app/internal/progress"
	"app/internal/repo"
	"app/internal/storage/block"
	"app/internal/storage/snapshot"

	"app/internal/util"
	"fmt"
	"path/filepath"
)

// ScanRepositoryBlocks verifies all repository blocks with a progress bar and returns an error if any are missing/damaged.
func ScanRepositoryBlocks() error {
	out, errCh := ScanRepositoryBlocksStream()

	totalBlocks, err := repo.CountAllBlocks()
	if err != nil {
		return err
	}

	bar := progress.NewProgress(totalBlocks, "Checking blocks")
	defer bar.Finish()

	for bc := range out {
		bar.Increment()
		if bc.Status != block.OK {
			return fmt.Errorf("block %s is missing or damaged", bc.Hash)
		}
	}

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}

// ScanRepositoryBlocksStream verifies blocks and streams results live.
func ScanRepositoryBlocksStream() (<-chan block.BlockCheck, <-chan error) {
	out := make(chan block.BlockCheck, 128)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		allBranches, err := core.Branches()
		if err != nil {
			errCh <- err
			return
		}

		type blockRef struct {
			files    map[string]struct{}
			branches map[string]struct{}
		}

		blockRefs := map[string]*blockRef{}
		blockHashes := map[string]struct{}{}

		// Phase 1: Collect all block references
		for _, branch := range allBranches {
			commitID, err := core.LastCommitID(branch.Name)
			if err != nil {
				errCh <- err
				return
			}
			if commitID == "" {
				continue
			}

			var commit core.Commit
			if err := util.ReadJSON(filepath.Join(config.CommitsDir, commitID+".json"), &commit); err != nil {
				continue
			}

			var fs snapshot.Fileset
			if err := util.ReadJSON(filepath.Join(config.FilesetsDir, commit.FilesetID+".json"), &fs); err != nil {
				continue
			}

			for _, f := range fs.Files {
				for _, b := range f.Blocks {
					r, ok := blockRefs[b.Hash]
					if !ok {
						r = &blockRef{
							files:    map[string]struct{}{},
							branches: map[string]struct{}{},
						}
						blockRefs[b.Hash] = r
						blockHashes[b.Hash] = struct{}{}
					}
					r.files[f.Path] = struct{}{}
					r.branches[branch.Name] = struct{}{}
				}
			}
		}

		// Phase 2: Verify blocks concurrently using storage API
		verifyOut := block.VerifyMany(blockHashes, 8)
		for bc := range verifyOut {
			ref := blockRefs[bc.Hash]
			bc.Files = util.SortedKeys(ref.files)
			bc.Branches = util.SortedKeys(ref.branches)
			out <- bc
		}
	}()

	return out, errCh
}
