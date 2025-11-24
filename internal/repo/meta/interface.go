package meta

type MetaContextInterface interface {
	GetCurrentBranch() (*Branch, error)
	GetBranch(name string) (Branch, error)
	ListBranches() ([]Branch, error)
	CreateBranch(name string) (Branch, error)
	BranchExists(name string) (bool, error)

	GetCommit(commitID string) (*Commit, error)
	CreateCommit(commit *Commit) (string, error)
	SetLastCommitID(branch, commitID string) error
	GetLastCommitID(branch string) (string, error)
	AllCommitIDs(branch string) ([]string, error)
	GetCommitsForBranch(branch string) ([]*Commit, error)
	GetLastCommitForBranch(branch string) (*Commit, error)

	GetHeadRef() (HeadRef, error)
	SetHeadRef(branch string) (HeadRef, error)
}
