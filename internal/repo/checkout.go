package repo

import (
	"fmt"

	"github.com/keshon/bvc/internal/repo/file"
)

type CheckoutResult struct {
	Branch   string
	CommitID string
	Empty    bool
}

func (r *Repository) Checkout(branchName string) (*CheckoutResult, error) {
	// ensure branch exists
	targetBranch, err := r.Meta.GetBranch(branchName)
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	// find target HEAD
	commitID, err := r.Meta.GetLastCommitID(targetBranch.Name)
	if err != nil {
		return nil, fmt.Errorf("load branch HEAD: %w", err)
	}

	// empty branch → simple cleanup + HEAD switch
	if commitID == "" {
		if err := r.checkoutApply(nil, branchName); err != nil {
			return nil, err
		}
		return &CheckoutResult{Branch: branchName, Empty: true}, nil
	}

	// load commit
	commit, err := r.Meta.GetCommit(commitID)
	if err != nil {
		return nil, fmt.Errorf("load commit %s: %w", commitID, err)
	}

	// load fileset
	nextFS, err := r.Snapshot.Load(commit.FilesetID)
	if err != nil {
		return nil, fmt.Errorf("load fileset %s: %w", commit.FilesetID, err)
	}

	// smart WT update
	if err := r.checkoutApply(nextFS.Files, branchName); err != nil {
		return nil, err
	}

	return &CheckoutResult{
		Branch:   branchName,
		CommitID: commitID,
		Empty:    false,
	}, nil
}

// checkoutApply updates WT to exactly match "targetFiles".
// This is a smart diff-based apply: add, update, delete minimal set.
func (r *Repository) checkoutApply(targetFiles []file.Entry, branchName string) error {
	// 1. scan current WT
	curr, err := r.Snapshot.BuildTrackedFileset()
	if err != nil {
		return fmt.Errorf("scan WT: %w", err)
	}

	// build lookup maps
	want := make(map[string]file.Entry, len(targetFiles))
	for _, f := range targetFiles {
		want[f.Path] = f
	}

	currMap := make(map[string]file.Entry, len(curr.Files))
	for _, f := range curr.Files {
		currMap[f.Path] = f
	}

	// 2. remove files that shouldn't exist anymore
	for path := range currMap {
		if _, ok := want[path]; !ok {
			if err := r.deleteFromWT(path); err != nil {
				return fmt.Errorf("delete %s: %w", path, err)
			}
		}
	}

	// 3. add/update files that should exist
	for path, entry := range want {
		currEntry, exists := currMap[path]

		if !exists || !entryEqual(currEntry, entry) {
			if err := r.File.Write(entry); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
	}

	// 4. update HEAD
	if _, err := r.Meta.SetHeadRef(branchName); err != nil {
		return fmt.Errorf("set HEAD: %w", err)
	}

	return nil
}

func entryEqual(a, b file.Entry) bool {
	if len(a.Blocks) != len(b.Blocks) {
		return false
	}
	for i := range a.Blocks {
		if a.Blocks[i].Hash != b.Blocks[i].Hash {
			return false
		}
	}
	return true
}

func (r *Repository) deleteFromWT(path string) error {
	return r.FS.Remove(path)
}
