package repotools

import (
	"app/internal/config"
	"app/internal/fsio"
	"app/internal/progress"
	"app/internal/repo"
	"app/internal/repo/store"
	"app/internal/repo/store/block"
	"app/internal/util"
	"fmt"
	"os"
)

// VerifyBlocks checks all blocks in repository and shows a progress bar.
// If onlyLatestCommit is false, collects blocks from all commits in all branches; otherwise only latest commits.
// Returns error if any block is missing/damaged.
func VerifyBlocks(r *repo.Repository, cfg *config.RepoConfig, onlyLatestCommit bool) error {
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
func VerifyBlocksStream(r *repo.Repository, cfg *config.RepoConfig, onlyLatestCommit bool) (<-chan block.BlockCheck, <-chan error) {
	out := make(chan block.BlockCheck, 128)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if _, err := fsio.StatFile(cfg.RepoRoot); os.IsNotExist(err) {
			errCh <- fmt.Errorf("repository not initialized (missing %s)", cfg.RepoRoot)
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

		st, err := store.NewStore(cfg)
		if err != nil {
			errCh <- fmt.Errorf("failed to init store: %w", err)
			return
		}
		verifyOut := st.Blocks.Verify(hashes, util.WorkerCount())

		for bc := range verifyOut {
			ref := blocks[bc.Hash]
			bc.Files = util.SortedKeys(ref.Files)
			bc.Branches = util.SortedKeys(ref.Branches)
			out <- bc
		}
	}()

	return out, errCh
}
