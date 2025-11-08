package store

import (
	"fmt"

	"app/internal/config"
	"app/internal/fsio"
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
	"app/internal/repo/store/snapshot"
)

// StoreContext is the high-level store abstraction that unifies all subsystems.
type StoreContext struct {
	Config    *config.RepoConfig
	Blocks    *block.BlockContext
	Files     *file.FileContext
	Snapshots *snapshot.SnapshotContext
}

// NewStore ensures that the store layout exists for the given repository config.
// It creates missing directories if necessary and returns a ready-to-use manager.
func NewStore(cfg *config.RepoConfig) (*StoreContext, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil RepoConfig provided")
	}

	if !isStoreExists(cfg) {
		if err := createStoreStructure(cfg); err != nil {
			return nil, err
		}
	}

	return buildManager(cfg), nil
}

// buildManager wires up the store subsystems.
func buildManager(cfg *config.RepoConfig) *StoreContext {
	st := &StoreContext{Config: cfg}

	st.Blocks = &block.BlockContext{
		Root: cfg.ObjectsDir(),
	}

	st.Files = &file.FileContext{
		Root:   cfg.WorkingTreeRoot,
		Blocks: st.Blocks,
	}

	st.Snapshots = &snapshot.SnapshotContext{
		Root:   cfg.FilesetsDir(),
		Files:  st.Files,
		Blocks: st.Blocks,
	}

	return st
}

// createRepoStructure builds the required directory structure if missing.
func createStoreStructure(cfg *config.RepoConfig) error {
	dirs := []string{
		cfg.CommitsDir(),
		cfg.FilesetsDir(),
		cfg.BranchesDir(),
		cfg.ObjectsDir(),
	}

	for _, d := range dirs {
		if err := fsio.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create store dir %q: %w", d, err)
		}
	}
	return nil
}

// isStoreExists verifies whether store directories exist.
func isStoreExists(cfg *config.RepoConfig) bool {
	// Minimal check: objects and commits dirs should exist
	return fsio.Exists(cfg.ObjectsDir()) && fsio.Exists(cfg.CommitsDir())
}
