package repo

import (
	"fmt"

	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/block"
	"github.com/keshon/bvc/internal/repo/file"
	"github.com/keshon/bvc/internal/repo/meta"
	"github.com/keshon/bvc/internal/repo/scan"
	"github.com/keshon/bvc/internal/repo/snapshot"
)

type Repository struct {
	Config   *config.RepoConfig
	Meta     meta.MetaContextInterface
	Block    block.BlockContextInterface
	File     file.FileContextInterface
	Snapshot snapshot.SnapshotContextInterface
}

func NewRepositoryByPath(path string) (*Repository, error) {
	cfg := config.NewRepoConfig(path)
	return NewRepository(cfg)
}

func NewRepository(cfg *config.RepoConfig) (*Repository, error) {
	// Create FS
	fs := fs.FS(&fs.OSFS{})

	// Create Meta
	meta, err := meta.NewMeta(cfg, fs)
	if err != nil {
		return nil, err
	}

	// Create BlockContext
	block := block.NewBlockContext(cfg.BlocksDir(), fs)

	// Create FileContext
	file := file.NewFileContext(cfg.WorkingTreeDir, cfg.RepoDir, block, fs)

	// Create file scanner
	scanner := scan.NewScanner(cfg.WorkingTreeDir, meta, fs)

	// Create SnapshotContext
	snapshot := snapshot.NewSnapshotContext(cfg.SnapshotsDir(), file, block, scanner, fs)

	// Ensure store layout
	if !isStoreExists(cfg, fs) {
		if err := createStoreStructure(cfg, fs); err != nil {
			return nil, err
		}
	}

	r := &Repository{
		Config:   cfg,
		Meta:     meta,
		Block:    block,
		File:     file,
		Snapshot: snapshot,
	}
	return r, nil
}

func (r *Repository) GetCommittedFileset(commitID string) (*snapshot.Fileset, error) {
	commit, err := r.Meta.GetCommit(commitID)
	if err != nil {
		return nil, err
	}
	fs, err := r.Snapshot.Load(commit.FilesetID)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func IsRepoExists(path string) bool {
	cfg := config.NewRepoConfig(path)
	return meta.IsMetaExists(cfg)
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
