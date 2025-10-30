package core

import (
	"app/internal/config"
	"app/internal/storage/snapshot"
	"app/internal/util"
	"errors"
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

// GetCommit gets a commit by its ID
// Returns commit object and error
func GetCommit(commitID string) (*Commit, error) {
	var c Commit
	path := filepath.Join(config.CommitsDir, commitID+".json")

	if err := util.ReadJSON(path, &c); err != nil {
		return nil, fmt.Errorf("failed to read commit %q: %w", commitID, err)
	}

	return &c, nil
}

// CreateCommit creates a new commit object
// Returns commit ID and error
func CreateCommit(commit *Commit) (string, error) {
	path := filepath.Join(config.CommitsDir, commit.ID+".json")

	if err := util.WriteJSON(path, commit); err != nil {
		return "", fmt.Errorf("failed to write commit %q: %w", commit.ID, err)
	}

	return commit.ID, nil
}

// SetLastCommit sets the last commit ID of the given branch
// Returns error
func SetLastCommitID(branch, commitID string) error {
	path := filepath.Join(config.BranchesDir, branch)
	if err := os.WriteFile(path, []byte(commitID), 0o644); err != nil {
		return fmt.Errorf("failed to set last commit for branch %q: %w", branch, err)
	}
	return nil
}

// GetLastCommitID returns the last commit ID of the given branch and error
func GetLastCommitID(branch string) (string, error) {
	path := filepath.Join(config.BranchesDir, branch)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read last commit for branch %q: %w", branch, err)
	}
	return string(data), nil
}

// AllCommitIDs returns all commit IDs for the given branch (latest to oldest)
func AllCommitIDs(branch string) ([]string, error) {
	lastID, err := GetLastCommitID(branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get last commit ID for branch %q: %w", branch, err)
	}
	if lastID == "" {
		return nil, nil
	}

	var ids []string
	seen := map[string]bool{}

	for id := lastID; id != ""; {
		if seen[id] {
			break
		}
		seen[id] = true
		ids = append(ids, id)

		c, err := GetCommit(id)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit %q: %w", id, err)
		}

		if len(c.Parents) == 0 || c.Parents[0] == "" {
			break
		}

		id = c.Parents[0]
	}

	return ids, nil
}

// GetCommitsForBranch gets all commits for a given branch
// Returns a slice of commit objects and an error
func GetCommitsForBranch(branch string) ([]*Commit, error) {
	ids, err := AllCommitIDs(branch)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	commits := make([]*Commit, 0, len(ids))
	for _, id := range ids {
		c, err := GetCommit(id)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit %q: %w", id, err)
		}
		commits = append(commits, c)
	}

	return commits, nil
}

// GetCommitFileset gets a commit's fileset by its ID
// Returns fileset object and error
func GetCommitFileset(commitID string) (*snapshot.Fileset, error) {
	commit, err := GetCommit(commitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %q: %w", commitID, err)
	}

	fs, err := snapshot.GetFileset(commit.FilesetID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("fileset %q not found for commit %q: %w", commit.FilesetID, commitID, err)
		}
		return nil, fmt.Errorf("failed to get fileset for commit %q: %w", commitID, err)
	}

	return &fs, nil
}
