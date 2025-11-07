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

// RepoConfig represents the on-disk configuration.
type RepoConfig struct {
	HashFormat string `json:"hash"`
}

// Repository represents an initialized repository.
type Repository struct {
	Config  *config.RepoConfig // replaces manual dir paths
	Meta    RepoConfig         // hash format etc.
	Storage *storage.Manager
}

// NewRepository constructs a Repository using the given root or default.
func NewRepository(root string) *Repository {
	cfg := config.NewRepoConfig(root)
	return &Repository{Config: cfg}
}

// InitAt initializes a repository at the provided path.
// Returns (*Repository, created, error).
func InitAt(path string, algo string) (*Repository, bool, error) {
	if algo == "" {
		algo = config.DefaultHash
	}

	r := NewRepository(path)
	cfg := r.Config

	// Detect existing repo
	if fi, err := fsio.StatFile(cfg.HeadFile()); err == nil && fi.Mode().IsRegular() {
		return r, false, os.ErrExist
	}

	// Create directories
	dirs := []string{
		cfg.Root,
		cfg.CommitsDir(),
		cfg.FilesetsDir(),
		cfg.BranchesDir(),
		cfg.ObjectsDir(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, false, fmt.Errorf("failed to create dir %q: %w", d, err)
		}
	}

	// Create default branch file
	mainBranch := filepath.Join(cfg.BranchesDir(), config.DefaultBranch)
	if err := fsio.WriteFile(mainBranch, []byte(""), 0o644); err != nil {
		return nil, false, fmt.Errorf("failed to create default branch: %w", err)
	}

	// Write HEAD
	headContent := "ref: branches/" + config.DefaultBranch
	if err := fsio.WriteFile(cfg.HeadFile(), []byte(headContent), 0o644); err != nil {
		return nil, false, fmt.Errorf("failed to write HEAD: %w", err)
	}

	// Write hash format to config file
	r.Config.HashFormat = algo
	if err := r.Config.SaveHash(); err != nil {
		return nil, false, fmt.Errorf("failed to save config.json: %w", err)
	}

	// Initialize storage manager
	st, err := storage.InitAt(cfg.Root)
	if err != nil {
		return nil, false, fmt.Errorf("failed to init storage manager: %w", err)
	}
	r.Storage = st

	return r, true, nil
}

// OpenAt opens an existing repository.
func OpenAt(path string) (*Repository, error) {
	r := NewRepository(path)
	r.Config.HashFormat = r.Config.GetSelectedHashName()
	cfg := r.Config

	if _, err := fsio.StatFile(cfg.HeadFile()); err != nil {
		return nil, fmt.Errorf("not a repository (missing HEAD): %w", err)
	}

	data, err := fsio.ReadFile(cfg.ConfigFile())
	if err != nil {
		return nil, fmt.Errorf("failed to read repo config: %w", err)
	}

	if err := json.Unmarshal(data, &r.Meta); err != nil {
		return nil, fmt.Errorf("failed to parse repo config: %w", err)
	}

	r.Storage, err = storage.OpenAt(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	return r, nil
}
