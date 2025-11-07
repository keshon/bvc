package repotools

import (
	"app/internal/config"
	"app/internal/fsio"
	"app/internal/progress"
	"app/internal/storage"
	"os"

	"app/internal/storage/block"

	"app/internal/util"
	"fmt"
)

// VerifyBlocks checks all blocks in repository and shows a progress bar.
// If onlyLatestCommit is false, collects blocks from all commits in all branches; otherwise only latest commits.
// Returns error if any block is missing/damaged.
func VerifyBlocks(r Repository, cfg *config.RepoConfig, onlyLatestCommit bool) error {
	out, errCh := VerifyBlocksStream(r, cfg, onlyLatestCommit)
	total, err := CountBlocks(r, cfg, onlyLatestCommit)
	if err != nil {
		return err
	}

	bar := progress.NewProgress(total, "Checking blocks")
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

// VerifyBlocksStream streams block verification results.
// If onlyLatestCommit is false, collects blocks from all commits in all branches; otherwise only latest commits.
// Returns error if any block is missing/damaged.
func VerifyBlocksStream(r Repository, cfg *config.RepoConfig, onlyLatestCommit bool) (<-chan block.BlockCheck, <-chan error) {
	out := make(chan block.BlockCheck, 128)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if _, err := fsio.StatFile(cfg.Root); os.IsNotExist(err) {
			errCh <- fmt.Errorf("repository not initialized (missing %s)", cfg.Root)
			return
		}

		blocks, err := ListAllBlocks(r, cfg, onlyLatestCommit)
		if err != nil {
			errCh <- err
			return
		}

		hashes := make(map[string]struct{}, len(blocks))
		for h := range blocks {
			hashes[h] = struct{}{}
		}

		mgr := storage.NewManager(cfg)
		verifyOut := mgr.Blocks.Verify(hashes, util.WorkerCount())

		for bc := range verifyOut {
			ref := blocks[bc.Hash]
			bc.Files = util.SortedKeys(ref.Files)
			bc.Branches = util.SortedKeys(ref.Branches)
			out <- bc
		}
	}()

	return out, errCh
}
