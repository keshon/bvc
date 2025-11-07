package repotools

import (
	"path/filepath"

	"app/internal/config"
	"app/internal/repo"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// ListAllBlocks returns a map[hash] of BlockInfo for all blocks in all branches.
// If onlyLatestCommit is false, collects blocks from all commits in all branches; otherwise only latest commits.
func ListAllBlocks(r Repository, onlyLatestCommit bool) (map[string]*BlockInfo, error) {
	branches, err := r.ListBranches()
	if err != nil {
		return nil, err
	}

	blocks := make(map[string]*BlockInfo)

	for _, b := range branches {
		var commitIDs []string
		if !onlyLatestCommit {
			commitIDs, err = r.AllCommitIDs(b.Name)
			if err != nil {
				return nil, err
			}
		} else {
			last, err := r.GetLastCommitID(b.Name)
			if err != nil {
				return nil, err
			}
			if last != "" {
				commitIDs = []string{last}
			}
		}

		for _, commitID := range commitIDs {
			commitPath := filepath.Join(config.CommitsDir, commitID+".json")
			var commit repo.Commit
			if err := util.ReadJSON(commitPath, &commit); err != nil {
				// skip missing commit file, but not silently fail everything
				continue
			}

			if commit.FilesetID == "" {
				continue
			}

			filesetPath := filepath.Join(config.FilesetsDir, commit.FilesetID+".json")
			var fs snapshot.Fileset
			if err := util.ReadJSON(filesetPath, &fs); err != nil {
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
// If onlyLatestCommit is false, counts blocks from all commits; otherwise only latest commits.
func CountBlocks(r Repository, onlyLatestCommit bool) (int, error) {
	branches, err := r.ListBranches()
	if err != nil {
		return 0, err
	}

	hashes := map[string]struct{}{}

	for _, b := range branches {
		var commitIDs []string
		if !onlyLatestCommit {
			commitIDs, err = r.AllCommitIDs(b.Name)
			if err != nil {
				return 0, err
			}
		} else {
			last, err := r.GetLastCommitID(b.Name)
			if err != nil {
				return 0, err
			}
			if last != "" {
				commitIDs = []string{last}
			}
		}

		for _, commitID := range commitIDs {
			commitPath := filepath.Join(config.CommitsDir, commitID+".json")
			var commit repo.Commit
			if err := util.ReadJSON(commitPath, &commit); err != nil {
				continue
			}

			if commit.FilesetID == "" {
				continue
			}

			filesetPath := filepath.Join(config.FilesetsDir, commit.FilesetID+".json")
			var fs snapshot.Fileset
			if err := util.ReadJSON(filesetPath, &fs); err != nil {
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
