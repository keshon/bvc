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
	RepoDir        string // repository root directory (absolute or relative)
	WorkingTreeDir string // working tree root directory
}

// NewRepoConfig creates a RepoConfig for a given root path.
// If root == "", it automatically resolves it using ResolveRepoDir().
func NewRepoConfig(root string) *RepoConfig {
	if root == "" {
		root = ResolveRepoDir()
	}
	cfg := &RepoConfig{RepoDir: root}
	cfg.WorkingTreeDir = ResolveWorkingTreeRoot()
	return cfg
}

// Derived Path Helpers
func (c *RepoConfig) RepoPath(parts ...string) string {
	return filepath.Join(append([]string{c.RepoDir}, parts...)...)
}

func (c *RepoConfig) CommitsDir() string {
	return c.RepoPath("commits")
}

func (c *RepoConfig) SnapshotsDir() string {
	return c.RepoPath("snapshots")
}

func (c *RepoConfig) BranchesDir() string {
	return c.RepoPath("branches")
}

func (c *RepoConfig) BlocksDir() string {
	return c.RepoPath("blocks")
}

func (c *RepoConfig) HeadFile() string {
	return c.RepoPath("HEAD")

}
