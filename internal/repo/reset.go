// repo/reset.go
package repo

import "fmt"

type ResetMode string

const (
	ResetSoft  ResetMode = "soft"
	ResetMixed ResetMode = "mixed"
	ResetHard  ResetMode = "hard"
)

func (r *Repository) Reset(commitID string, mode ResetMode) error {
	// resolve current branch
	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	// if no commitID, just clear index (soft/mixed default behavior)
	if commitID == "" {
		return r.File.ClearIndex()
	}

	// validate commit exists
	commit, err := r.Meta.GetCommit(commitID)
	if err != nil {
		return fmt.Errorf("unknown commit: %s", commitID)
	}

	// move HEAD
	if err := r.Meta.SetLastCommitID(branch.Name, commitID); err != nil {
		return err
	}

	// reset index if mixed or hard
	if mode == ResetMixed || mode == ResetHard {
		if err := r.resetIndex(commit.FilesetID); err != nil {
			return err
		}
	}

	// reset working tree if hard
	if mode == ResetHard {
		if err := r.resetWorkingTree(commit.FilesetID); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) resetIndex(filesetID string) error {
	fs, err := r.Snapshot.Load(filesetID)
	if err != nil {
		return err
	}

	if err := r.File.ClearIndex(); err != nil {
		return err
	}
	return r.File.SaveIndexReplace(fs.Files)
}

func (r *Repository) resetWorkingTree(filesetID string) error {
	fs, err := r.Snapshot.Load(filesetID)
	if err != nil {
		return err
	}
	return r.File.RestoreFilesToWorkingTree(fs.Files, "files")
}
