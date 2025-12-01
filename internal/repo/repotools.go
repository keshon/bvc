package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/keshon/bvc/internal/progress"
	"github.com/keshon/bvc/internal/repo/block"
	"github.com/keshon/bvc/internal/repo/meta"
	"github.com/keshon/bvc/internal/repo/snapshot"
	"github.com/keshon/bvc/internal/util"
)

type BlockInfo struct {
	Size     int64
	Files    map[string]struct{}
	Branches map[string]struct{}
}

// CountBlocks returns the total number of blocks in all branches.
// If onlyLatestCommit is false, counts blocks from all commits; otherwise only latest commits.
func (r *Repository) CountBlocks(onlyLatestCommit bool) (int, error) {
	branches, err := r.Meta.ListBranches()
	if err != nil {
		return 0, err
	}

	hashes := map[string]struct{}{}

	for _, b := range branches {
		var commitIDs []string
		if !onlyLatestCommit {
			commitIDs, err = r.Meta.AllCommitIDs(b.Name)
		} else {
			last, _ := r.Meta.GetLastCommitID(b.Name)
			if last != "" {
				commitIDs = []string{last}
			}
		}
		if err != nil {
			return 0, err
		}

		for _, commitID := range commitIDs {
			commitPath := filepath.Join(r.Config.CommitsDir(), commitID+".json")
			var commit meta.Commit
			if err := r.readJSON(commitPath, &commit); err != nil {
				continue
			}
			if commit.FilesetID == "" {
				continue
			}

			filesetPath := filepath.Join(r.Config.SnapshotsDir(), commit.FilesetID+".json")
			var fs snapshot.Fileset
			if err := r.readJSON(filesetPath, &fs); err != nil {
				continue
			}

			for _, file := range fs.Files {
				for _, blk := range file.Blocks {
					hashes[blk.Hash] = struct{}{}
				}
			}
		}
	}

	return len(hashes), nil
}

// ListAllBlocks returns a map[hash]*BlockInfo for all blocks in all branches.
// cfg defines the repository root (e.g., config.NewRepoConfig(".bvc")).
func (r *Repository) ListAllBlocks(onlyLatestCommit bool) (map[string]*BlockInfo, error) {
	branches, err := r.Meta.ListBranches()
	if err != nil {
		return nil, err
	}

	blocks := make(map[string]*BlockInfo)

	for _, b := range branches {
		var commitIDs []string
		if !onlyLatestCommit {
			commitIDs, err = r.Meta.AllCommitIDs(b.Name)
		} else {
			var last string
			last, err = r.Meta.GetLastCommitID(b.Name)
			if err == nil && last != "" {
				commitIDs = []string{last}
			}
		}
		if err != nil {
			return nil, err
		}

		for _, commitID := range commitIDs {
			commitPath := filepath.Join(r.Config.CommitsDir(), commitID+".json")
			var commit meta.Commit
			if err := r.readJSON(commitPath, &commit); err != nil {
				continue
			}
			if commit.FilesetID == "" {
				continue
			}

			filesetPath := filepath.Join(r.Config.SnapshotsDir(), commit.FilesetID+".json")
			var fs snapshot.Fileset
			if err := r.readJSON(filesetPath, &fs); err != nil {
				continue
			}

			for _, f := range fs.Files {
				for _, blk := range f.Blocks {
					info, ok := blocks[blk.Hash]
					if !ok {
						info = &BlockInfo{
							Size:     blk.Size,
							Files:    map[string]struct{}{},
							Branches: map[string]struct{}{},
						}
						blocks[blk.Hash] = info
					}
					info.Files[f.Path] = struct{}{}
					info.Branches[b.Name] = struct{}{}
				}
			}
		}
	}

	return blocks, nil
}

// VerifyBlocks checks all blocks in repository and shows a progress bar.
// If onlyLatestCommit is false, collects blocks from all commits in all branches; otherwise only latest commits.
// Returns error if any block is missing/damaged.
func (r *Repository) VerifyBlocks(onlyLatestCommit bool) error {
	out, errCh := r.VerifyBlocksStream(onlyLatestCommit)
	total, err := r.CountBlocks(onlyLatestCommit)
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
func (r *Repository) VerifyBlocksStream(onlyLatestCommit bool) (<-chan block.BlockCheck, <-chan error) {
	out := make(chan block.BlockCheck, 128)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		if _, err := r.FS.Stat(r.Config.RepoDir); os.IsNotExist(err) {
			errCh <- fmt.Errorf("repository not initialized (missing %s)", r.Config.RepoDir)
			return
		}

		blocks, err := r.ListAllBlocks(onlyLatestCommit)
		if err != nil {
			errCh <- err
			return
		}

		hashes := make(map[string]struct{}, len(blocks))
		for h := range blocks {
			hashes[h] = struct{}{}
		}

		verifyOut := r.Block.Verify(hashes, util.WorkerCount())

		for bc := range verifyOut {
			ref := blocks[bc.Hash]
			bc.Files = util.SortedKeys(ref.Files)
			bc.Branches = util.SortedKeys(ref.Branches)
			out <- bc
		}
	}()

	return out, errCh
}
