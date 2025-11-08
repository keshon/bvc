package meta

import (
	"fmt"
	"path/filepath"

	"app/internal/config"
	"app/internal/fsio"
)

// MetaContext represents an initialized repository.
type MetaContext struct {
	Config *config.RepoConfig
}

// NewMeta ensures a repository exists at the given root.
// It will create all necessary structure if missing.
// Returns an initialized MetaContext or an error.
func NewMeta(cfg *config.RepoConfig) (*MetaContext, error) {
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

	return &MetaContext{
		Config: cfg,
	}, nil
}

// createMetaStructure builds a fresh meta layout and writes defaults.
func createMetaStructure(cfg *config.RepoConfig) error {
	dirs := []string{
		cfg.RepoRoot,
		cfg.CommitsDir(),
		cfg.FilesetsDir(),
		cfg.BranchesDir(),
		cfg.ObjectsDir(),
	}
	for _, d := range dirs {
		if err := fsio.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create dir %q: %w", d, err)
		}
	}

	mainBranch := filepath.Join(cfg.BranchesDir(), config.DefaultBranch)
	if err := fsio.WriteFile(mainBranch, []byte(""), 0o644); err != nil {
		return fmt.Errorf("failed to create default branch: %w", err)
	}

	headContent := "ref: branches/" + config.DefaultBranch
	if err := fsio.WriteFile(cfg.HeadFile(), []byte(headContent), 0o644); err != nil {
		return fmt.Errorf("failed to write HEAD: %w", err)
	}

	return nil
}

// isMetaExists checks if the given meta config points to an existing meta.
func IsMetaExists(cfg *config.RepoConfig) bool {
	fi, err := fsio.StatFile(cfg.HeadFile())
	return err == nil && fi.Mode().IsRegular()
}
