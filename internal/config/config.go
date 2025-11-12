package config

import (
	"path/filepath"
)

// Constants

var (
	IsDev               = false
	RepoDir             = ".bvc"
	RepoPointerFile     = ".bvc-pointer"
	IgnoredFilesFile    = ".bvc-ignore"
	DefaultBranch       = "main"
	DefaultIgnoredFiles = []string{RepoPointerFile, RepoDir}
)

// RepoConfig represents a resolved repository configuration and layout.
type RepoConfig struct {
	RepoRoot        string // repository root directory (absolute or relative)
	WorkingTreeRoot string // working tree root directory
}

// NewRepoConfig creates a RepoConfig for a given root path.
// If root == "", it automatically resolves it using ResolveRepoRoot().
func NewRepoConfig(root string) *RepoConfig {
	if root == "" {
		root = ResolveRepoRoot()
	}
	cfg := &RepoConfig{RepoRoot: root}
	cfg.WorkingTreeRoot = ResolveWorkingTreeRoot()
	return cfg
}

// Derived Path Helpers
func (c *RepoConfig) RepoPath(parts ...string) string {
	return filepath.Join(append([]string{c.RepoRoot}, parts...)...)
}

func (c *RepoConfig) CommitsDir() string {
	return c.RepoPath("commits")
}

func (c *RepoConfig) FilesetsDir() string {
	return c.RepoPath("filesets")
}

func (c *RepoConfig) BranchesDir() string {
	return c.RepoPath("branches")
}

func (c *RepoConfig) ObjectsDir() string {
	return c.RepoPath("objects")
}

func (c *RepoConfig) HeadFile() string {
	return c.RepoPath("HEAD")

}
