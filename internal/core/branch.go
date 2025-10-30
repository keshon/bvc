package core

import (
	"app/internal/config"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Branch struct {
	Name string
}

// GetCurrentBranch gets the current branch object and an error
func GetCurrentBranch() (Branch, error) {
	ref, err := GetHeadRef()
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get HEAD ref: %w", err)
	}

	name := filepath.Base(ref.String())
	if name == "" {
		return Branch{}, fmt.Errorf("HEAD ref is empty or invalid")
	}

	return Branch{Name: name}, nil
}

// GetBranch gets a branch
// Returns a branch object and an error
func GetBranch(name string) (Branch, error) {
	// check if branch exists
	exist, err := branchExists(name)
	if err != nil {
		return Branch{}, fmt.Errorf("failed to check if branch exists: %w", err)
	}
	if !exist {
		return Branch{}, fmt.Errorf("branch '%s' does not exist", name)
	}
	return Branch{Name: name}, nil
}

// GetBranches gets all branches
// Returns a slice of branch objects and an error
func GetBranches() ([]Branch, error) {
	dirEntries, err := os.ReadDir(config.BranchesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read branches directory: %w", err)
	}

	branches := make([]Branch, 0, len(dirEntries))
	for _, entry := range dirEntries {
		branches = append(branches, Branch{Name: entry.Name()})
	}

	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})

	return branches, nil
}

// CreateBranch creates a new branch and sets last commit of parent branch
// Returns a branch object and an error
func CreateBranch(name string) (Branch, error) {
	branch, err := GetCurrentBranch()
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get current branch: %w", err)
	}

	currCommit, err := GetLastCommitID(branch.Name)
	if err != nil {
		return Branch{}, fmt.Errorf("failed to get last commit ID: %w", err)
	}

	path := filepath.Join(config.BranchesDir, name)
	if _, err := os.Stat(path); err == nil {
		return Branch{}, fmt.Errorf("branch '%s' already exists: %w", name, os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Branch{}, fmt.Errorf("failed to check branch existence: %w", err)
	}

	if err := os.WriteFile(path, []byte(currCommit), 0o644); err != nil {
		return Branch{}, fmt.Errorf("could not save new branch: %w", err)
	}

	return Branch{Name: name}, nil
}

// branchExists checks if a branch exists
func branchExists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(config.BranchesDir, name))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat branch: %w", err)
}
