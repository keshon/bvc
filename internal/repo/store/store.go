package store

import (
	"fmt"

	"app/internal/config"
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

// NewStoreOptions allows optional dependency injection (FS, BlockStore)
type NewStoreOptions struct {
	FS     file.FS
	Blocks *block.BlockContext
}

// NewStoreDefault creates a store with default dependencies (FS, BlockStore)
func NewStoreDefault(cfg *config.RepoConfig) (*StoreContext, error) {
	return NewStore(cfg, nil)
}

// NewStore creates a store with optional dependencies (FS, BlockStore)
func NewStore(cfg *config.RepoConfig, opts *NewStoreOptions) (*StoreContext, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil RepoConfig provided")
	}

	fs := file.FS(&file.OSFS{}) // default FS
	if opts != nil && opts.FS != nil {
		fs = opts.FS
	}

	blocks := &block.BlockContext{
		Root: cfg.ObjectsDir(),
	}
	if opts != nil && opts.Blocks != nil {
		blocks = opts.Blocks
	}

	// Ensure store layout
	if !isStoreExists(cfg, fs) {
		if err := createStoreStructure(cfg, fs); err != nil {
			return nil, err
		}
	}

	return &StoreContext{
		Config: cfg,
		Blocks: blocks,
		Files: &file.FileContext{
			Root:     cfg.WorkingTreeRoot,
			RepoRoot: cfg.RepoRoot,
			Blocks:   blocks,
			FS:       fs,
		},
		Snapshots: &snapshot.SnapshotContext{
			Root:   cfg.FilesetsDir(),
			Files:  &file.FileContext{Root: cfg.WorkingTreeRoot, RepoRoot: cfg.RepoRoot, Blocks: blocks, FS: fs},
			Blocks: blocks,
		},
	}, nil
}

// createStoreStructure builds required dirs via injected FS
func createStoreStructure(cfg *config.RepoConfig, fs file.FS) error {
	dirs := []string{
		cfg.CommitsDir(),
		cfg.FilesetsDir(),
		cfg.BranchesDir(),
		cfg.ObjectsDir(),
	}

	for _, d := range dirs {
		if err := fs.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create store dir %q: %w", d, err)
		}
	}
	return nil
}

// isStoreExists uses FS to verify directories
func isStoreExists(cfg *config.RepoConfig, fs file.FS) bool {
	return exists(fs, cfg.ObjectsDir()) && exists(fs, cfg.CommitsDir())
}

func exists(fs file.FS, path string) bool {
	info, err := fs.Stat(path)
	return err == nil && info.IsDir()
}
