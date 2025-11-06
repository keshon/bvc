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
// Returns error if any block is missing/damaged.
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

// VerifyBlocksStream streams block verification results.
func VerifyBlocksStream(allHistory bool) (<-chan block.BlockCheck, <-chan error) {
	out := make(chan block.BlockCheck, 128)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		// Open the repo and access its storage manager
		mgr, err := storage.NewManager(config.ResolveRepoRoot()), error(nil)
		if _, statErr := fsio.StatFile(config.ResolveRepoRoot()); os.IsNotExist(statErr) {
			errCh <- fmt.Errorf("%s", "repository not initialized (missing "+config.ResolveRepoRoot()+")")
			return
		}

		// Collect all referenced blocks
		blocks, err := ListAllBlocks(allHistory)
		if err != nil {
			errCh <- err
			return
		}

		// Prepare hash set
		hashes := make(map[string]struct{}, len(blocks))
		for h := range blocks {
			hashes[h] = struct{}{}
		}

		// Use the block subsystem under the manager
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
