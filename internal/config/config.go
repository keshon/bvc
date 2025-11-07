package config

import (
	"app/internal/fsio"
	"encoding/json"
	"path/filepath"
)

const IsDev = false

var (
	RepoDir         = ".bvc"
	RepoPointerFile = ".bvc-pointer"
	CommitsDir      = "commits"
	FilesetsDir     = "filesets"
	BranchesDir     = "branches"
	ObjectsDir      = "objects"
	HeadFile        = "HEAD"
)

const (
	DefaultBranch = "main"
	DefaultHash   = "xxh3" // "xxh3" | "sha256"
)

var DefaultIgnoredFiles = []string{RepoPointerFile, RepoDir}
var RepoRootOverride string

// GetSelectedHashName returns the configured hash algorithm (e.g. "xxh3", "blake3", etc.).
// Falls back to "xxh3" if not specified or config is missing.
func GetSelectedHashName() string {
	cfgPath := filepath.Join(ResolveRepoRoot(), "config.json")

	data, err := fsio.ReadFile(cfgPath)
	if err != nil {
		return DefaultHash
	}

	var cfg struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultHash
	}
	if cfg.Hash == "" {
		return DefaultHash
	}
	return cfg.Hash
}

// ResolveRepoRoot returns the actual repository root, respecting .bvc-pointer or .bvc directory.
func ResolveRepoRoot() string {
	if RepoRootOverride != "" {
		return RepoRootOverride
	}
	root := RepoDir

	// Check if pointer file exists
	if fi, err := fsio.StatFile(RepoPointerFile); err == nil && !fi.IsDir() {
		data, err := fsio.ReadFile(RepoPointerFile)
		if err == nil {
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
