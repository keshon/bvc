package meta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/keshon/bvc/internal/util"
)

type Commit struct {
	ID        string   `json:"id"`
	Parents   []string `json:"parents"`
	Branch    string   `json:"branch"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	FilesetID string   `json:"fileset_id"`
}

// GetCommit reads a commit by ID.
func (mc *MetaContext) GetCommit(commitID string) (*Commit, error) {
	var c Commit
	path := filepath.Join(mc.Config.CommitsDir(), commitID+".json")
	if err := util.ReadJSON(path, &c); err != nil {
		return nil, fmt.Errorf("failed to read commit %q: %w", commitID, err)
	}
	return &c, nil
}

// CreateCommit writes a commit to store.
func (mc *MetaContext) CreateCommit(commit *Commit) (string, error) {
	path := filepath.Join(mc.Config.CommitsDir(), commit.ID+".json")
	if err := util.WriteJSON(path, commit); err != nil {
		return "", fmt.Errorf("failed to write commit %q: %w", commit.ID, err)
	}
	return commit.ID, nil
}

// SetLastCommitID writes the branch last-commit pointer.
func (mc *MetaContext) SetLastCommitID(branch, commitID string) error {
	path := filepath.Join(mc.Config.BranchesDir(), branch)
	if err := mc.FS.WriteFile(path, []byte(commitID), 0o644); err != nil {
		return fmt.Errorf("failed to set last commit for branch %q: %w", branch, err)
	}
	return nil
}

// GetLastCommitID returns the last commit ID for branch.
func (mc *MetaContext) GetLastCommitID(branch string) (string, error) {
	path := filepath.Join(mc.Config.BranchesDir(), branch)
	data, err := mc.FS.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read last commit for branch %q: %w", branch, err)
	}
	return string(data), nil
}

// AllCommitIDs returns all commit IDs for branch (latest -> oldest).
func (mc *MetaContext) AllCommitIDs(branch string) ([]string, error) {
	lastID, err := mc.GetLastCommitID(branch)
	if err != nil {
		return nil, err
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

		c, err := mc.GetCommit(id)
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

// GetCommitsForBranch returns []*Commit (latest -> oldest).
func (mc *MetaContext) GetCommitsForBranch(branch string) ([]*Commit, error) {
	ids, err := mc.AllCommitIDs(branch)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	commits := make([]*Commit, 0, len(ids))
	for _, id := range ids {
		c, err := mc.GetCommit(id)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit %q: %w", id, err)
		}
		commits = append(commits, c)
	}
	return commits, nil
}

// GetLastCommitForBranch returns the last commit for branch.
func (mc *MetaContext) GetLastCommitForBranch(branch string) (*Commit, error) {
	lastID, err := mc.GetLastCommitID(branch)
	if err != nil {
		return nil, err
	}
	if lastID == "" {
		return nil, nil
	}
	return mc.GetCommit(lastID)
}
