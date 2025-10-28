package repo

import (
	"app/internal/config"
	"app/internal/core"
	"app/internal/progress"

	"app/internal/storage/block"
	"app/internal/storage/snapshot"

	"app/internal/util"
	"fmt"
	"path/filepath"
)

// VerifyBlocks verifies all repository blocks with a progress bar and returns an error if any are missing/damaged.
func VerifyBlocks(allHistory bool) error {
	out, errCh := VerifyBlocksStream(allHistory)

	totalBlocks, err := CountBlocks(allHistory)
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

// VerifyBlocksStream verifies blocks and streams results live.// VerifyBlocksStream verifies blocks and streams results live.
// If allHistory is true, it collects blocks from all commits in all branches; otherwise only latest commits.
func VerifyBlocksStream(allHistory bool) (<-chan block.BlockCheck, <-chan error) {
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

		type blockMeta struct {
			files    map[string]struct{}
			branches map[string]struct{}
		}

		blockMetas := map[string]*blockMeta{}
		blockHashes := map[string]struct{}{}

		for _, branch := range allBranches {
			var commitIDs []string
			if allHistory {
				commitIDs, err = core.AllCommitIDs(branch.Name)
				if err != nil {
					errCh <- err
					return
				}
			} else {
				last, err := core.LastCommitID(branch.Name)
				if err != nil {
					errCh <- err
					return
				}
				if last != "" {
					commitIDs = []string{last}
				}
			}

			for _, commitID := range commitIDs {
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
						r, ok := blockMetas[b.Hash]
						if !ok {
							r = &blockMeta{
								files:    map[string]struct{}{},
								branches: map[string]struct{}{},
							}
							blockMetas[b.Hash] = r
							blockHashes[b.Hash] = struct{}{}
						}
						r.files[f.Path] = struct{}{}
						r.branches[branch.Name] = struct{}{}
					}
				}
			}
		}

		verifyOut := block.VerifyBlocks(blockHashes, 8)
		for bc := range verifyOut {
			ref := blockMetas[bc.Hash]
			bc.Files = util.SortedKeys(ref.files)
			bc.Branches = util.SortedKeys(ref.branches)
			out <- bc
		}
	}()

	return out, errCh
}
