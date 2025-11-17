package repotools

import (
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"

	"github.com/keshon/bvc/internal/progress"

	"fmt"
	"os"

	"github.com/keshon/bvc/internal/repo/store"
	"github.com/keshon/bvc/internal/repo/store/block"
	"github.com/keshon/bvc/internal/util"
)

// VerifyBlocks checks all blocks in repository and shows a progress bar.
// If onlyLatestCommit is false, collects blocks from all commits in all branches; otherwise only latest commits.
// Returns error if any block is missing/damaged.
func VerifyBlocks(m MetaInterface, cfg *config.RepoConfig, onlyLatestCommit bool) error {
	out, errCh := VerifyBlocksStream(m, cfg, onlyLatestCommit)
	total, err := CountBlocks(m, cfg, onlyLatestCommit)
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
func VerifyBlocksStream(m MetaInterface, cfg *config.RepoConfig, onlyLatestCommit bool) (<-chan block.BlockCheck, <-chan error) {
	fs := fs.NewOSFS()
	out := make(chan block.BlockCheck, 128)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if _, err := fs.Stat(cfg.RepoDir); os.IsNotExist(err) {
			errCh <- fmt.Errorf("repository not initialized (missing %s)", cfg.RepoDir)
			return
		}

		blocks, err := ListAllBlocks(m, cfg, onlyLatestCommit)
		if err != nil {
			errCh <- err
			return
		}

		hashes := make(map[string]struct{}, len(blocks))
		for h := range blocks {
			hashes[h] = struct{}{}
		}

		st, err := store.NewStoreDefault(cfg)
		if err != nil {
			errCh <- fmt.Errorf("failed to init store: %w", err)
			return
		}
		verifyOut := st.BlockCtx.Verify(hashes, util.WorkerCount())

		for bc := range verifyOut {
			ref := blocks[bc.Hash]
			bc.Files = util.SortedKeys(ref.Files)
			bc.Branches = util.SortedKeys(ref.Branches)
			out <- bc
		}
	}()

	return out, errCh
}
