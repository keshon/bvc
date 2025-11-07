package repotools

import "app/internal/repo"

// Repository is a subset of repo.Repository
type Repository interface {
	ListBranches() ([]repo.Branch, error)
	AllCommitIDs(branch string) ([]string, error)
	GetLastCommitID(branch string) (string, error)
}

// BlockInfo holds metadata about a block in the repository
type BlockInfo struct {
	Size     int64
	Files    map[string]struct{}
	Branches map[string]struct{}
}
