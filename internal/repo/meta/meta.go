package meta

import (
	"fmt"

	"path/filepath"

	"app/internal/config"
	"app/internal/fs"
)

// MetaContext represents an initialized repository.
type MetaContext struct {
	Config *config.RepoConfig
	FS     fs.FS
}

// NewMetaDefault creates a meta with default dependencies (FS)
// and ensures a repository exists at the given root.
func NewMetaDefault(cfg *config.RepoConfig) (*MetaContext, error) {
	return NewMeta(cfg, nil)
}

// NewMeta ensures a repository exists at the given root.
// It will create all necessary structure if missing.
// Returns an initialized MetaContext or an error.
func NewMeta(cfg *config.RepoConfig, targetFS fs.FS) (*MetaContext, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil MetaConfig provided")
	}

	// open existing meta if valid
	if IsMetaExists(cfg) {
		return &MetaContext{Config: cfg}, nil
	}

	// create new meta structure if missing
	if err := createMetaStructure(cfg); err != nil {
		return nil, err
	}

	// init fs
	if targetFS == nil {
		targetFS = fs.NewOSFS()
	}

	return &MetaContext{
		Config: cfg,
		FS:     targetFS,
	}, nil
}

// createMetaStructure builds a fresh meta layout and writes defaults.
func createMetaStructure(cfg *config.RepoConfig) error {
	fs := fs.NewOSFS()

	dirs := []string{
		cfg.RepoRoot,
		cfg.CommitsDir(),
		cfg.SnapshotsDir(),
		cfg.BranchesDir(),
		cfg.BlocksDir(),
	}
	for _, d := range dirs {
		if err := fs.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create dir %q: %w", d, err)
		}
	}

	mainBranch := filepath.Join(cfg.BranchesDir(), config.DefaultBranch)
	if err := fs.WriteFile(mainBranch, []byte(""), 0o644); err != nil {
		return fmt.Errorf("failed to create default branch: %w", err)
	}

	headContent := "ref: branches/" + config.DefaultBranch
	if err := fs.WriteFile(cfg.HeadFile(), []byte(headContent), 0o644); err != nil {
		return fmt.Errorf("failed to write HEAD: %w", err)
	}

	return nil
}

// isMetaExists checks if the given meta config points to an existing meta.
func IsMetaExists(cfg *config.RepoConfig) bool {
	fs := fs.NewOSFS()
	fi, err := fs.Stat(cfg.HeadFile())
	return err == nil && fi.Mode().IsRegular()
}
