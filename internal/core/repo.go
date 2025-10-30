package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"app/internal/config"
)

// Repository represents an initialized repository at a given path.
// Prefer creating/opening via InitAt/OpenAt (or Init/Open which use config.RepoDir).
type Repository struct {
	Path        string // base repo directory, e.g. ".bvc"
	CommitsDir  string
	FilesetsDir string
	BranchesDir string
	ObjectsDir  string
	HeadFile    string
}

// NewRepository constructs a Repository pointing at path.
// It does NOT check the filesystem.
func NewRepository(path string) *Repository {
	return &Repository{
		Path:        path,
		CommitsDir:  filepath.Join(path, config.CommitsDir),
		FilesetsDir: filepath.Join(path, config.FilesetsDir),
		BranchesDir: filepath.Join(path, config.BranchesDir),
		ObjectsDir:  filepath.Join(path, config.ObjectsDir),
		HeadFile:    filepath.Join(path, config.HeadFile),
	}
}

// InitAt initializes a repository at the provided path.
// Returns (*Repository, created, error).
// - created=true when the repo did not exist and was created by this call.
// - created=false when the repo already existed (idempotent).
func InitAt(path string) (*Repository, bool, error) {
	r := NewRepository(path)
	// If the repo already exists and has HEAD -> not created.
	if fi, err := os.Stat(r.Path); err == nil && fi.IsDir() {
		if _, err := os.Stat(r.HeadFile); err == nil {
			return r, false, fmt.Errorf("repository already initialized at %q", r.Path)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, false, fmt.Errorf("failed to stat HEAD file: %w", err)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, false, fmt.Errorf("failed to stat repo dir %q: %w", r.Path, err)
	}

	// Create directories (idempotent)
	dirs := []string{
		r.Path,
		r.CommitsDir,
		r.FilesetsDir,
		r.BranchesDir,
		r.ObjectsDir,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, false, fmt.Errorf("failed to create directory %q: %w", d, err)
		}
	}

	// Ensure default branch file exists and HEAD points at it.
	mainHeadPath := filepath.Join(r.BranchesDir, config.DefaultBranch)
	if _, err := os.Stat(mainHeadPath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(mainHeadPath, []byte(""), 0o644); err != nil {
			return nil, false, fmt.Errorf("failed to create default branch file %q: %w", mainHeadPath, err)
		}
		headContent := "ref: branches/" + config.DefaultBranch
		if err := os.WriteFile(r.HeadFile, []byte(headContent), 0o644); err != nil {
			return nil, false, fmt.Errorf("failed to write HEAD file %q: %w", r.HeadFile, err)
		}
	} else if err != nil {
		return nil, false, fmt.Errorf("failed to stat default branch file %q: %w", mainHeadPath, err)
	}

	return r, true, nil
}

// OpenAt opens an existing repository (validates HEAD exists).
func OpenAt(path string) (*Repository, error) {
	r := NewRepository(path)
	if _, err := os.Stat(r.Path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("repository not found at %q", r.Path)
		}
		return nil, fmt.Errorf("failed to stat repo dir %q: %w", r.Path, err)
	}
	if _, err := os.Stat(r.HeadFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("not a repository (missing HEAD) at %q", r.Path)
		}
		return nil, fmt.Errorf("failed to stat HEAD file %q: %w", r.HeadFile, err)
	}
	return r, nil
}

// Get root folder where repo dir is stored.
func (r *Repository) Root() string {
	// we need return the folder name that is a parent of the repo folder, not the repo folder itself
	// if .bvc is in /temp/.bvc, we need to return /temp
	return filepath.Dir(r.Path)
}
