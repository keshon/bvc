package repo

import (
	"app/internal/fsio"
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
func (r *Repository) GetCurrentBranch() (Branch, error) {
	ref, err := r.GetHeadRef()
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get HEAD ref: %w", err)
	}
	name := filepath.Base(ref.String())
	if name == "" {
		return Branch{}, fmt.Errorf("HEAD ref is empty or invalid")
	}
	return Branch{Name: name}, nil
}

// GetBranch returns a Branch if it exists.
func (r *Repository) GetBranch(name string) (Branch, error) {
	exists, err := r.BranchExists(name)
	if err != nil {
		return Branch{}, fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return Branch{}, fmt.Errorf("branch %q does not exist", name)
	}
	return Branch{Name: name}, nil
}

// ListBranches returns all branches sorted by name.
func (r *Repository) ListBranches() ([]Branch, error) {
	dirEntries, err := fsio.ReadDir(r.BranchesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read branches directory %q: %w", r.BranchesDir, err)
	}
	branches := make([]Branch, 0, len(dirEntries))
	for _, e := range dirEntries {
		branches = append(branches, Branch{Name: e.Name()})
	}
	sort.Slice(branches, func(i, j int) bool { return branches[i].Name < branches[j].Name })
	return branches, nil
}

// CreateBranch creates a new branch pointing at the current HEAD commit.
func (r *Repository) CreateBranch(name string) (Branch, error) {
	curr, err := r.GetCurrentBranch()
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get current branch: %w", err)
	}
	lastID, err := r.GetLastCommitID(curr.Name)
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get last commit ID: %w", err)
	}

	path := filepath.Join(r.BranchesDir, name)
	if _, err := fsio.StatFile(path); err == nil {
		return Branch{}, fmt.Errorf("branch %q already exists: %w", name, os.ErrExist)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Branch{}, fmt.Errorf("failed to check branch file %q: %w", path, err)
	}

	if err := fsio.WriteFile(path, []byte(lastID), 0o644); err != nil {
		return Branch{}, fmt.Errorf("failed to write branch file %q: %w", path, err)
	}
	return Branch{Name: name}, nil
}

// BranchExists checks for branch existence (fast).
func (r *Repository) BranchExists(name string) (bool, error) {
	_, err := fsio.StatFile(filepath.Join(r.BranchesDir, name))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat branch file: %w", err)
}
