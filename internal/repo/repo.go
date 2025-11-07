package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"app/internal/config"
	"app/internal/fsio"
	"app/internal/storage"
)

// Repository represents an initialized repository.
type Repository struct {
	Config  *config.RepoConfig // replaces manual dir paths
	Storage *storage.StorageManager
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

	// Detect existing repo
	if fi, err := fsio.StatFile(r.Config.HeadFile()); err == nil && fi.Mode().IsRegular() {
		return r, false, os.ErrExist
	}

	// Create directories
	dirs := []string{
		r.Config.Root,
		r.Config.CommitsDir(),
		r.Config.FilesetsDir(),
		r.Config.BranchesDir(),
		r.Config.ObjectsDir(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, false, fmt.Errorf("failed to create dir %q: %w", d, err)
		}
	}

	// Create default branch file
	mainBranch := filepath.Join(r.Config.BranchesDir(), config.DefaultBranch)
	if err := fsio.WriteFile(mainBranch, []byte(""), 0o644); err != nil {
		return nil, false, fmt.Errorf("failed to create default branch: %w", err)
	}

	// Write HEAD
	headContent := "ref: branches/" + config.DefaultBranch
	if err := fsio.WriteFile(r.Config.HeadFile(), []byte(headContent), 0o644); err != nil {
		return nil, false, fmt.Errorf("failed to write HEAD: %w", err)
	}

	// Write hash format to config file
	r.Config.HashFormat = algo
	if err := r.Config.SaveHash(); err != nil {
		return nil, false, fmt.Errorf("failed to save config.json: %w", err)
	}

	// Initialize storage manager
	st, err := storage.InitAt(r.Config.Root)
	if err != nil {
		return nil, false, fmt.Errorf("failed to init storage manager: %w", err)
	}
	r.Storage = st

	return r, true, nil
}

// OpenAt opens an existing repository.
func OpenAt(path string) (*Repository, error) {
	r := NewRepository(path)
	r.Config.HashFormat = r.Config.GetHash()
	cfg := r.Config

	if _, err := fsio.StatFile(cfg.HeadFile()); err != nil {
		return nil, fmt.Errorf("not a repository (missing HEAD): %w", err)
	}

	var err error
	r.Storage, err = storage.OpenAt(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	return r, nil
}
