package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const IsDev = false

const (
	RepoDir     = ".bvc"
	CommitsDir  = "commits"
	FilesetsDir = "filesets"
	BranchesDir = "branches"
	ObjectsDir  = "objects"
	HeadFile    = "HEAD"

	BVCPointerFile = ".bvc-pointer"
)

const (
	DefaultBranch = "main"
)

const (
	DefaultHash = "xxh3" // "xxh3" | "sha256"
)

var IgnoredFiles = []string{BVCPointerFile, RepoDir}

// SelectedHash returns the configured hash algorithm (e.g. "xxh3", "blake3", etc.).
// Falls back to "xxh3" if not specified or config is missing.
func SelectedHash() string {
	cfgPath := filepath.Join(DetectRepoRoot(), "config.json")

	data, err := os.ReadFile(cfgPath)
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

// DetectRepoRoot returns the actual repository root, respecting .bvc-pointer.
func DetectRepoRoot() string {
	root := RepoDir

	// Check if pointer file exists
	if fi, err := os.Stat(BVCPointerFile); err == nil && !fi.IsDir() {
		data, err := os.ReadFile(BVCPointerFile)
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
