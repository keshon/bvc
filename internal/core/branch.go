package core

import (
	"app/internal/config"
	"fmt"
	"os"
	"path/filepath"
)

type Branch struct {
	Name string
}

// CurrentBranch returns the name of the current branch
func CurrentBranch() (Branch, error) {
	ref, err := GetHeadRef()
	if err != nil {
		return Branch{}, err
	}
	return Branch{Name: filepath.Base(ref.String())}, nil
}

// Branches returns ordered list of branches
func Branches() ([]Branch, error) {
	dirEntries, err := os.ReadDir(config.BranchesDir)
	if err != nil {
		return nil, err
	}

	branches := make([]Branch, 0, len(dirEntries))
	for _, entry := range dirEntries {
		branches = append(branches, Branch{Name: entry.Name()})
	}

	for i := 0; i < len(branches)-1; i++ {
		for j := i + 1; j < len(branches); j++ {
			if branches[i].Name > branches[j].Name {
				branches[i], branches[j] = branches[j], branches[i]
			}
		}
	}

	return branches, nil
}

func IsBranchExist(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(config.BranchesDir, name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetBranch(name string) (Branch, error) {
	//check if branch exists
	exist, err := IsBranchExist(name)
	if err != nil {
		return Branch{}, err
	}
	if !exist {
		return Branch{}, fmt.Errorf("branch '%s' does not exist", name)
	}
	return Branch{Name: name}, nil
}

// CreateBranch creates a new branch and sets last commit of parent branch
func CreateBranch(name string) (Branch, error) {
	branch, err := CurrentBranch()
	if err != nil {
		return Branch{}, err
	}
	currCommit, err := LastCommitID(branch.Name)
	if err != nil {
		return Branch{}, err
	}
	path := filepath.Join(config.BranchesDir, name)
	if _, err := os.Stat(path); err == nil {
		return Branch{}, fmt.Errorf("branch already exists")
	}
	err = os.WriteFile(path, []byte(currCommit), 0644)
	if err != nil {
		return Branch{}, fmt.Errorf("could save a new branch")
	}
	return Branch{Name: name}, nil
}
