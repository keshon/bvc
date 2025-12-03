package repo

import (
	"fmt"
	"time"

	"github.com/keshon/bvc/internal/repo/meta"
)

func (r *Repository) CherryPick(commitID string) (*meta.Commit, error) {
	// load target commit
	targetCommit, err := r.Meta.GetCommit(commitID)
	if err != nil {
		return nil, fmt.Errorf("load commit: %w", err)
	}

	// load its fileset
	targetFileset, err := r.Snapshot.Load(targetCommit.FilesetID)
	if err != nil {
		return nil, fmt.Errorf("load fileset: %w", err)
	}

	// resolve current branch
	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("get current branch: %w", err)
	}

	// find last commit on branch
	parentID, err := r.Meta.GetLastCommitID(branch.Name)
	if err != nil {
		return nil, fmt.Errorf("load branch HEAD: %w", err)
	}

	// construct new commit
	newCommit := &meta.Commit{
		ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
		Parents:   []string{parentID},
		Branch:    branch.Name,
		Message:   fmt.Sprintf("Pick commit %s", commitID),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: targetCommit.FilesetID,
	}

	// write commit
	if _, err := r.Meta.CreateCommit(newCommit); err != nil {
		return nil, fmt.Errorf("write commit: %w", err)
	}

	// update branch pointer
	if err := r.Meta.SetLastCommitID(branch.Name, newCommit.ID); err != nil {
		return nil, fmt.Errorf("update branch pointer: %w", err)
	}

	// restore working tree
	if err := r.File.RestoreFilesToWorkingTree(
		targetFileset.Files,
		fmt.Sprintf("cherry-pick %s", commitID),
	); err != nil {
		return nil, fmt.Errorf("restore WT: %w", err)
	}

	return newCommit, nil
}
