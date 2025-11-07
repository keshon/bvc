package repotools

import (
	"path/filepath"

	"app/internal/config"
	"app/internal/repo"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// CountBlocks returns the total number of blocks in all branches.
// If onlyLatestCommit is false, counts blocks from all commits; otherwise only latest commits.
func CountBlocks(r Repository, cfg *config.RepoConfig, onlyLatestCommit bool) (int, error) {
	branches, err := r.ListBranches()
	if err != nil {
		return 0, err
	}

	hashes := map[string]struct{}{}

	for _, b := range branches {
		var commitIDs []string
		if !onlyLatestCommit {
			commitIDs, err = r.AllCommitIDs(b.Name)
		} else {
			last, _ := r.GetLastCommitID(b.Name)
			if last != "" {
				commitIDs = []string{last}
			}
		}
		if err != nil {
			return 0, err
		}

		for _, commitID := range commitIDs {
			commitPath := filepath.Join(cfg.CommitsDir(), commitID+".json")
			var commit repo.Commit
			if err := util.ReadJSON(commitPath, &commit); err != nil {
				continue
			}

			if commit.FilesetID == "" {
				continue
			}

			filesetPath := filepath.Join(cfg.FilesetsDir(), commit.FilesetID+".json")
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
