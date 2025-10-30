package repo

import (
	"fmt"
	"path/filepath"

	"app/internal/config"
	"app/internal/core"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// BlockInfo holds metadata about a block in the repository
type BlockInfo struct {
	Size     int64
	Files    map[string]struct{}
	Branches map[string]struct{}
}

// ListAllBlocks returns a map[hash] of BlockInfo for all blocks in all branches.
// If allHistory is true, collects blocks from all commits in all branches; otherwise only latest commits.
func ListAllBlocks(allHistory bool) (map[string]*BlockInfo, error) {
	branches, err := core.GetBranches()
	if err != nil {
		return nil, err
	}

	blocks := make(map[string]*BlockInfo)

	for _, b := range branches {
		var commitIDs []string
		if allHistory {
			commitIDs, err = core.AllCommitIDs(b.Name)
			if err != nil {
				return nil, err
			}
		} else {
			last, err := core.GetLastCommitID(b.Name)
			if err != nil {
				return nil, err
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

// CountBlocks returns the total number of blocks in all branches.
// If allHistory is true, counts blocks from all commits; otherwise only latest commits.
func CountBlocks(allHistory bool) (int, error) {
	branches, err := core.GetBranches()
	if err != nil {
		return 0, err
	}
	hashes := map[string]struct{}{}

	for _, b := range branches {
		var commitIDs []string
		if allHistory {
			commitIDs, err = core.AllCommitIDs(b.Name)
			if err != nil {
				return 0, err
			}
		} else {
			last, err := core.GetLastCommitID(b.Name)
			if err != nil {
				return 0, err
			}
			if last != "" {
				commitIDs = []string{last}
			}
		}

		for _, commitID := range commitIDs {
			var commit core.Commit
			if err := util.ReadJSON(fmt.Sprintf("%s/%s.json", config.CommitsDir, commitID), &commit); err != nil {
				continue
			}
			var fs snapshot.Fileset
			if err := util.ReadJSON(fmt.Sprintf("%s/%s.json", config.FilesetsDir, commit.FilesetID), &fs); err != nil {
				continue
			}
			for _, e := range fs.Files {
				for _, b := range e.Blocks {
					hashes[b.Hash] = struct{}{}
				}
			}
		}
	}

	return len(hashes), nil
}
