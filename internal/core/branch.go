package core

import (
	"app/internal/config"
	"fmt"
	"os"
	"path/filepath"
)

// CurrentBranch returns the name of the current branch
func CurrentBranch() (string, error) {
	head, err := HeadRef()
	if err != nil {
		return "", err
	}
	return filepath.Base(head), nil
}

// Branches returns a list of branch names
func Branches() ([]string, error) {
	dirEntries, err := os.ReadDir(config.BranchesDir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(dirEntries))
	for _, entry := range dirEntries {
		names = append(names, entry.Name())
	}
	return names, nil
}

// CreateBranch creates a new branch and sets last commit of parent branch
func CreateBranch(name string) error {
	currBranch, err := CurrentBranch()
	if err != nil {
		return err
	}
	currCommit, err := LastCommit(currBranch)
	if err != nil {
		return err
	}
	path := filepath.Join(config.BranchesDir, name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("branch already exists")
	}
	return os.WriteFile(path, []byte(currCommit), 0644)
}
