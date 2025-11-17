package store

import (
	"fmt"

	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/store/block"
	"github.com/keshon/bvc/internal/repo/store/file"
	"github.com/keshon/bvc/internal/repo/store/snapshot"
)

// StoreContext is the high-level store abstraction that unifies all subsystems.
type StoreContext struct {
	Config      *config.RepoConfig
	BlockCtx    *block.BlockContext
	FileCtx     *file.FileContext
	SnapshotCtx *snapshot.SnapshotContext
}

// NewStoreOptions allows optional dependency injection (FS, BlockStore)
type NewStoreOptions struct {
	FS          fs.FS
	BlockCtx    *block.BlockContext
	FileCtx     *file.FileContext
	SnapshotCtx *snapshot.SnapshotContext
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

	// Resolve FS
	fs := fs.FS(&fs.OSFS{})
	if opts != nil && opts.FS != nil {
		fs = opts.FS
	}

	// Resolve BlockContext
	blockCtx := block.NewBlockContext(cfg.BlocksDir(), fs)
	if opts != nil && opts.BlockCtx != nil {
		blockCtx = opts.BlockCtx
	}

	// Resolve FileContext
	fileCtx := file.NewFileContext(cfg.WorkingTreeDir, cfg.RepoDir, blockCtx, fs)
	if opts != nil && opts.FileCtx != nil {
		fileCtx = opts.FileCtx
	}

	// Resolve SnapshotContext
	snapshotCtx := snapshot.NewSnapshotContext(cfg.SnapshotsDir(), fileCtx, blockCtx, fs)
	if opts != nil && opts.SnapshotCtx != nil {
		snapshotCtx = opts.SnapshotCtx
	}

	// Ensure store layout
	if !isStoreExists(cfg, fs) {
		if err := createStoreStructure(cfg, fs); err != nil {
			return nil, err
		}
	}

	return &StoreContext{
		Config:      cfg,
		BlockCtx:    blockCtx,
		FileCtx:     fileCtx,
		SnapshotCtx: snapshotCtx,
	}, nil
}

// createStoreStructure builds required dirs via injected FS
func createStoreStructure(cfg *config.RepoConfig, fs fs.FS) error {
	dirs := []string{
		cfg.CommitsDir(),
		cfg.SnapshotsDir(),
		cfg.BranchesDir(),
		cfg.BlocksDir(),
	}

	for _, d := range dirs {
		if err := fs.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create store dir %q: %w", d, err)
		}
	}
	return nil
}

// isStoreExists uses FS to verify directories
func isStoreExists(cfg *config.RepoConfig, fs fs.FS) bool {
	return exists(fs, cfg.BlocksDir()) && exists(fs, cfg.CommitsDir())
}

func exists(fs fs.FS, path string) bool {
	info, err := fs.Stat(path)
	return err == nil && info.IsDir()
}
