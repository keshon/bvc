package config

const IsDev = false

const (
	RepoDir     = ".bvc"
	CommitsDir  = RepoDir + "/commits"
	FilesetsDir = RepoDir + "/filesets"
	BranchesDir = RepoDir + "/branches"
	ObjectsDir  = RepoDir + "/objects"
)

const (
	DefaultBranch = "main"
	HeadFile      = "HEAD"
)
