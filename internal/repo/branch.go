package repo

import (
	"fmt"
	"sort"
)

// CreateBranch creates a branch starting from a given start point.
// If startPoint == "", use HEAD.
func (r *Repository) CreateBranch(name, startPoint string, force bool) (string, error) {
	exists, err := r.Meta.BranchExists(name)
	if err != nil {
		return "", fmt.Errorf("check branch existence: %w", err)
	}

	if exists && !force {
		return "", fmt.Errorf("branch %q already exists", name)
	}

	// resolve start commit
	var commitID string

	if startPoint == "" {
		head, err := r.Meta.GetCurrentBranch()
		if err != nil {
			return "", fmt.Errorf("resolve HEAD: %w", err)
		}
		commitID, err = r.Meta.GetLastCommitID(head.Name)
		if err != nil {
			return "", fmt.Errorf("resolve HEAD commit: %w", err)
		}
	} else {
		// treat startPoint as commit or branch
		commit, err := r.resolveCommitish(startPoint)
		if err != nil {
			return "", fmt.Errorf("resolve %q: %w", startPoint, err)
		}
		commitID = commit
	}

	// create branch with commit pointer
	br, err := r.Meta.CreateBranchAt(name, commitID, force)
	if err != nil {
		return "", fmt.Errorf("create branch: %w", err)
	}

	return br.Name, nil
}

// DeleteBranch removes a branch. Safe delete means the branch must be merged into current branch.
func (r *Repository) DeleteBranch(name string, force bool) error {
	current, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	if name == current.Name {
		return fmt.Errorf("cannot delete current branch %q", name)
	}

	// ensure the branch exists
	_, err = r.Meta.GetBranch(name)
	if err != nil {
		return fmt.Errorf("branch %q does not exist: %w", name, err)
	}

	if !force {
		// check if fully merged
		merged, err := r.isBranchFullyMerged(name, current.Name)
		if err != nil {
			return err
		}
		if !merged {
			return fmt.Errorf("branch %q is not fully merged; use -D to force", name)
		}
	}

	return r.Meta.DeleteBranch(name)
}

// Rename branch (safe and forced)
func (r *Repository) RenameBranch(old, new string, force bool) error {
	if old == new {
		return fmt.Errorf("nothing to rename")
	}

	exists, err := r.Meta.BranchExists(new)
	if err != nil {
		return err
	}

	if exists && !force {
		return fmt.Errorf("branch %q already exists; use -M to overwrite", new)
	}

	return r.Meta.RenameBranch(old, new, force)
}

// ListBranches returns current branch and sorted list of all branches.
func (r *Repository) ListBranches() (string, []string, error) {
	cur, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return "", nil, fmt.Errorf("get current branch: %w", err)
	}

	all, err := r.Meta.ListBranches()
	if err != nil {
		return "", nil, fmt.Errorf("list branches: %w", err)
	}

	names := make([]string, len(all))
	for i, b := range all {
		names[i] = b.Name
	}

	sort.Strings(names)
	return cur.Name, names, nil
}

// resolveCommitish tries to resolve string to commit SHA.
func (r *Repository) resolveCommitish(name string) (string, error) {
	// try branch name
	if exists, _ := r.Meta.BranchExists(name); exists {
		return r.Meta.GetLastCommitID(name)
	}
	// try commit ID
	commit, err := r.Meta.GetCommit(name)
	if err == nil && commit != nil {
		return name, nil
	}

	return "", fmt.Errorf("unknown commit or branch %q", name)
}

// isBranchFullyMerged returns true if src is reachable from target.
func (r *Repository) isBranchFullyMerged(src, target string) (bool, error) {
	srcCommits, err := r.Meta.AllCommitIDs(src)
	if err != nil {
		return false, err
	}

	targetCommits, err := r.Meta.AllCommitIDs(target)
	if err != nil {
		return false, err
	}

	set := map[string]struct{}{}
	for _, id := range targetCommits {
		set[id] = struct{}{}
	}

	// if every commit in src is reachable from target - fully merged
	for _, id := range srcCommits {
		if _, ok := set[id]; !ok {
			return false, nil
		}
	}
	return true, nil
}
