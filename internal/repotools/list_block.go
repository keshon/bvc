package repotools

import (
	"path/filepath"

	"app/internal/config"
	"app/internal/repo"
	"app/internal/repo/meta"

	"app/internal/repo/store/snapshot"
	"app/internal/util"
)

// ListAllBlocks returns a map[hash]*BlockInfo for all blocks in all branches.
// cfg defines the repository root (e.g., config.NewRepoConfig(".bvc")).
func ListAllBlocks(r *repo.Repository, cfg *config.RepoConfig, onlyLatestCommit bool) (map[string]*BlockInfo, error) {
	branches, err := r.Meta.ListBranches()
	if err != nil {
		return nil, err
	}

	blocks := make(map[string]*BlockInfo)

	for _, b := range branches {
		var commitIDs []string
		var err error
		if !onlyLatestCommit {
			commitIDs, err = r.Meta.AllCommitIDs(b.Name)
		} else {
			var last string
			last, err = r.Meta.GetLastCommitID(b.Name) // capture error
			if err == nil && last != "" {
				commitIDs = []string{last}
			}
		}
		if err != nil {
			return nil, err // propagate error
		}

		for _, commitID := range commitIDs {
			commitPath := filepath.Join(cfg.CommitsDir(), commitID+".json")
			var commit meta.Commit
			if err := util.ReadJSON(commitPath, &commit); err != nil {
				continue // skip missing commit
			}
			if commit.FilesetID == "" {
				continue
			}

			filesetPath := filepath.Join(cfg.FilesetsDir(), commit.FilesetID+".json")
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
