package config

import (
	"app/internal/fsio"
	"encoding/json"
	"path/filepath"
)

// Constants

const (
	IsDev = false

	RepoDir         = ".bvc"
	RepoPointerFile = ".bvc-pointer"

	DefaultBranch = "main"
	DefaultHash   = "xxh3" // or "sha256"
)

var (
	Hashes              = []string{"xxh3", "sha256"}
	DefaultIgnoredFiles = []string{RepoPointerFile, RepoDir}
)

// RepoConfig represents a resolved repository configuration and layout.
type RepoConfig struct {
	Root       string // repository root directory (absolute or relative)
	HashFormat string `json:"hash"`
}

// NewRepoConfig creates a RepoConfig for a given root path.
// If root == "", it automatically resolves it using ResolveRepoRoot().
func NewRepoConfig(root string) *RepoConfig {
	if root == "" {
		root = ResolveRepoRoot()
	}
	return &RepoConfig{Root: filepath.Clean(root)}
}

// Derived Path Helpers
func (c *RepoConfig) RepoPath(parts ...string) string {
	return filepath.Join(append([]string{c.Root}, parts...)...)
}
func (c *RepoConfig) CommitsDir() string  { return c.RepoPath("commits") }
func (c *RepoConfig) FilesetsDir() string { return c.RepoPath("filesets") }
func (c *RepoConfig) BranchesDir() string { return c.RepoPath("branches") }
func (c *RepoConfig) ObjectsDir() string  { return c.RepoPath("objects") }
func (c *RepoConfig) HeadFile() string    { return c.RepoPath("HEAD") }
func (c *RepoConfig) ConfigFile() string  { return c.RepoPath("config.json") }

// Hash Config Loading

// GetSelectedHashName reads hash format from config.json in the repo root.
// Returns DefaultHash if not found or invalid.
func (c *RepoConfig) GetSelectedHashName() string {
	data, err := fsio.ReadFile(c.ConfigFile())
	if err != nil {
		return DefaultHash
	}

	var cfg struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.Hash == "" {
		return DefaultHash
	}
	return cfg.Hash
}

// SaveHash writes the current HashFormat to config.json in the repo.
func (c *RepoConfig) SaveHash() error {
	if c.HashFormat == "" {
		c.HashFormat = DefaultHash
	}
	data, err := json.MarshalIndent(struct {
		Hash string `json:"hash"`
	}{c.HashFormat}, "", "  ")
	if err != nil {
		return err
	}
	return fsio.WriteFile(c.ConfigFile(), data, 0o644)
}

// ResolveRepoRoot determines the actual repository root.
// It respects the .bvc-pointer file, if it exists.
func ResolveRepoRoot() string {
	root := RepoDir

	if fi, err := fsio.StatFile(RepoPointerFile); err == nil && !fi.IsDir() {
		if data, err := fsio.ReadFile(RepoPointerFile); err == nil {
			target := filepath.Clean(string(data))
			if filepath.IsAbs(target) {
				root = target
			} else {
				root = filepath.Join(".", target)
			}
		}
	}

	return root
}
