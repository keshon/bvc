package repotools

import "app/internal/repo/meta"

// MetaInterface is the minimal interface repotools needs from a repo's meta layer.
type MetaInterface interface {
	ListBranches() ([]meta.Branch, error)
	AllCommitIDs(branch string) ([]string, error)
	GetLastCommitID(branch string) (string, error)
}

// BlockInfo holds metadata about a block in the repository
type BlockInfo struct {
	Size     int64
	Files    map[string]struct{}
	Branches map[string]struct{}
}
