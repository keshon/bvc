package repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"app/internal/config"
	"app/internal/fsio"
	"app/internal/storage"
)

// RepoConfig represents the repository configuration saved on disk.
type RepoConfig struct {
	HashFormat string `json:"hash"`
}

// Repository represents an initialized repository.
type Repository struct {
	Root        string
	CommitsDir  string
	FilesetsDir string
	BranchesDir string
	ObjectsDir  string
	HeadFile    string
	ConfigFile  string

	Config  RepoConfig
	Storage *storage.Manager
}

// NewRepository constructs a Repository pointing at root directory.
func NewRepository(root string) (*Repository, error) {
	r := &Repository{
		Root:        filepath.Clean(root),
		CommitsDir:  filepath.Join(root, config.CommitsDir),
		FilesetsDir: filepath.Join(root, config.FilesetsDir),
		BranchesDir: filepath.Join(root, config.BranchesDir),
		ObjectsDir:  filepath.Join(root, config.ObjectsDir),
		HeadFile:    filepath.Join(root, config.HeadFile),
		ConfigFile:  filepath.Join(root, "config.json"),
	}
	return r, nil
}

// InitAt initializes a repository at the provided path.
// Returns (*Repository, created, error).
// - created=true when the repo did not exist and was created by this call.
// - created=false when the repo already existed (idempotent).
func InitAt(path string, algo string) (*Repository, bool, error) {
	if algo == "" {
		algo = config.DefaultHash
	}

	r, err := NewRepository(path)
	if err != nil {
		return nil, false, err
	}

	// Detect already-initialized repo
	if fi, err := fsio.StatFile(r.HeadFile); err == nil && fi.Mode().IsRegular() {
		return r, false, os.ErrExist
	}

	// Create directories
	dirs := []string{r.Root, r.CommitsDir, r.FilesetsDir, r.BranchesDir, r.ObjectsDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, false, fmt.Errorf("failed to create dir %q: %w", d, err)
		}
	}

	// Create default branch file
	mainBranch := filepath.Join(r.BranchesDir, config.DefaultBranch)
	if err := fsio.WriteFile(mainBranch, []byte(""), 0o644); err != nil {
		return nil, false, fmt.Errorf("failed to create default branch: %w", err)
	}

	// Write HEAD file
	headContent := "ref: branches/" + config.DefaultBranch
	if err := fsio.WriteFile(r.HeadFile, []byte(headContent), 0o644); err != nil {
		return nil, false, fmt.Errorf("failed to write HEAD: %w", err)
	}

	// Write repo config
	r.Config = RepoConfig{HashFormat: algo}
	data, _ := json.MarshalIndent(r.Config, "", "  ")
	if err := fsio.WriteFile(r.ConfigFile, data, 0o644); err != nil {
		return nil, false, fmt.Errorf("failed to write config.json: %w", err)
	}

	// Init storage manager
	r.Storage, err = storage.InitAt(path)
	if err != nil {
		return nil, false, fmt.Errorf("failed to init storage manager: %w", err)
	}

	return r, true, nil
}

// OpenAt opens an existing repository.
func OpenAt(path string) (*Repository, error) {
	r, err := NewRepository(path)
	if err != nil {
		return nil, err
	}

	if _, err := fsio.StatFile(r.HeadFile); err != nil {
		return nil, fmt.Errorf("not a repository (missing HEAD): %w", err)
	}

	data, err := fsio.ReadFile(r.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo config: %w", err)
	}
	if err := json.Unmarshal(data, &r.Config); err != nil {
		return nil, fmt.Errorf("failed to parse repo config: %w", err)
	}

	r.Storage, err = storage.InitAt(path)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	return r, nil
}
