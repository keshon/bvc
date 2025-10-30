package core

import (
	"app/internal/config"
	"app/internal/storage/snapshot"
	"app/internal/util"
	"fmt"
	"os"
	"path/filepath"
)

type Commit struct {
	ID        string   `json:"id"`
	Parents   []string `json:"parents"`
	Branch    string   `json:"branch"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	FilesetID string   `json:"fileset_id"`
}

// GetCommit returns the commit with the given ID
func GetCommit(commitID string) (*Commit, error) {
	var c Commit
	path := filepath.Join(config.CommitsDir, commitID+".json")
	if err := util.ReadJSON(path, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateCommit creates a new commit
func CreateCommit(commit *Commit) (string, error) {
	path := filepath.Join(config.CommitsDir, commit.ID+".json")
	err := util.WriteJSON(path, commit)
	if err != nil {
		return "", err
	}
	return commit.ID, nil
}

// SetLastCommit sets the last commit ID of the given branch
func SetLastCommitID(branch, commitID string) error {
	path := filepath.Join(config.BranchesDir, branch)
	return os.WriteFile(path, []byte(commitID), 0644)
}

// LastCommitID returns the last commit ID of the given branch
func LastCommitID(branch string) (string, error) {
	path := filepath.Join(config.BranchesDir, branch)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AllCommitIDs returns all commit IDs of the given branch
func AllCommitIDs(branch string) ([]string, error) {
	path := filepath.Join(config.BranchesDir, branch)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return []string{string(data)}, nil
}

// GetBranchCommmits returns a slice of commits for selected branch
func GetCommitsForBranch(branch string, fn func(*Commit) bool) error {
	commitID, err := LastCommitID(branch)
	if err != nil {
		return err
	}
	if commitID == "" {
		return nil
	}

	seen := map[string]bool{}

	for commitID != "" {
		if seen[commitID] {
			break
		}
		seen[commitID] = true

		c, err := GetCommit(commitID)
		if err != nil {
			return fmt.Errorf("read commit %s: %w", commitID, err)
		}

		if cont := fn(c); !cont {
			break
		}

		if len(c.Parents) == 0 {
			break
		}
		commitID = c.Parents[0]
	}

	return nil
}

// GetCommitFileset loads the Fileset object associated with a given commit ID.
func GetCommitFileset(commitID string) (*snapshot.Fileset, error) {
	commit, err := GetCommit(commitID)
	if err != nil {
		return nil, err
	}

	fs, err := snapshot.LoadFileset(commit.FilesetID)
	if err != nil {
		return nil, err
	}

	return &fs, nil
}
