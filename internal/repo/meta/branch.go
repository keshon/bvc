package meta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Branch represents a branch name.
type Branch struct {
	Name string
}

// GetCurrentBranch returns the current branch.
func (mc *MetaContext) GetCurrentBranch() (*Branch, error) {
	ref, err := mc.GetHeadRef()
	if err != nil {
		return &Branch{}, fmt.Errorf("failed to get HEAD ref: %w", err)
	}
	name := filepath.Base(ref.String())
	if name == "" {
		return &Branch{}, fmt.Errorf("HEAD ref is empty or invalid")
	}
	return &Branch{Name: name}, nil
}

// GetBranch returns a Branch if it exists.
func (mc *MetaContext) GetBranch(name string) (Branch, error) {
	exists, err := mc.BranchExists(name)
	if err != nil {
		return Branch{}, fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return Branch{}, fmt.Errorf("branch %q does not exist", name)
	}
	return Branch{Name: name}, nil
}

// ListBranches returns all branches sorted by name.
func (mc *MetaContext) ListBranches() ([]Branch, error) {
	dirEntries, err := mc.FS.ReadDir(mc.Config.BranchesDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read branches directory %q: %w", mc.Config.BranchesDir(), err)
	}
	branches := make([]Branch, 0, len(dirEntries))
	for _, e := range dirEntries {
		branches = append(branches, Branch{Name: e.Name()})
	}
	sort.Slice(branches, func(i, j int) bool { return branches[i].Name < branches[j].Name })
	return branches, nil
}

// CreateBranch creates a new branch pointing at the current HEAD commit.
// TODO: remove it or replace with CreateBranchAt
func (mc *MetaContext) CreateBranch(name string) (Branch, error) {
	curr, err := mc.GetCurrentBranch()
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get current branch: %w", err)
	}
	lastID, err := mc.GetLastCommitID(curr.Name)
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get last commit ID: %w", err)
	}

	path := filepath.Join(mc.Config.BranchesDir(), name)
	if _, err := mc.FS.Stat(path); err == nil {
		return Branch{}, fmt.Errorf("branch %q already exists: %w", name, os.ErrExist)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Branch{}, fmt.Errorf("failed to check branch file %q: %w", path, err)
	}

	if err := mc.FS.WriteFile(path, []byte(lastID), 0o644); err != nil {
		return Branch{}, fmt.Errorf("failed to write branch file %q: %w", path, err)
	}
	return Branch{Name: name}, nil
}

// BranchExists checks for branch existence (fast).
func (mc *MetaContext) BranchExists(name string) (bool, error) {
	_, err := mc.FS.Stat(filepath.Join(mc.Config.BranchesDir(), name))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat branch file: %w", err)
}

func (mc *MetaContext) CreateBranchAt(name, commitID string, force bool) (Branch, error) {
	// validate commitID if provided
	if commitID != "" {
		_, err := mc.GetCommit(commitID)
		if err != nil {
			return Branch{}, fmt.Errorf("invalid commit %q: %w", commitID, err)
		}
	}

	path := filepath.Join(mc.Config.BranchesDir(), name)

	_, err := mc.FS.Stat(path)
	if err == nil {
		if !force {
			return Branch{}, fmt.Errorf("branch %q already exists", name)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Branch{}, fmt.Errorf("failed to stat branch %q: %w", name, err)
	}

	// commitID "" means no commits yet (allowed)
	if err := mc.FS.WriteFile(path, []byte(commitID), 0o644); err != nil {
		return Branch{}, fmt.Errorf("failed to write branch %q: %w", name, err)
	}

	return Branch{Name: name}, nil
}

func (mc *MetaContext) DeleteBranch(name string) error {
	// can't delete current branch
	cur, err := mc.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	if cur.Name == name {
		return fmt.Errorf("cannot delete branch %q: currently checked out", name)
	}

	path := filepath.Join(mc.Config.BranchesDir(), name)

	_, err = mc.FS.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("branch %q does not exist", name)
	}
	if err != nil {
		return fmt.Errorf("failed to stat branch %q: %w", name, err)
	}

	if err := mc.FS.Remove(path); err != nil {
		return fmt.Errorf("failed to delete branch %q: %w", name, err)
	}

	return nil
}

func (mc *MetaContext) RenameBranch(old, new string, force bool) error {
	if old == new {
		return nil
	}

	oldPath := filepath.Join(mc.Config.BranchesDir(), old)
	newPath := filepath.Join(mc.Config.BranchesDir(), new)

	// must exist
	if _, err := mc.FS.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("branch %q does not exist", old)
	} else if err != nil {
		return fmt.Errorf("failed to stat branch %q: %w", old, err)
	}

	// new branch validation
	_, err := mc.FS.Stat(newPath)
	if err == nil && !force {
		return fmt.Errorf("branch %q already exists", new)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to stat branch %q: %w", new, err)
	}

	// move branch ref file
	if err := mc.FS.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename branch %q -> %q: %w", old, new, err)
	}

	// update HEAD if needed
	cur, err := mc.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	if cur.Name == old {
		if _, err := mc.SetHeadRef(new); err != nil {
			return fmt.Errorf("failed to update HEAD to %q: %w", new, err)
		}
	}

	return nil
}
