package repo

import (
	"encoding/json"
	"fmt"
	"path/filepath"

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
	FS       fs.FS
}

func NewRepositoryByPath(path string) (*Repository, error) {
	cfg := config.NewRepoConfig(path)
	return NewRepositoryWithFS(cfg, nil) // nil triggers OSFS
}

func NewRepository(cfg *config.RepoConfig) (*Repository, error) {
	return NewRepositoryWithFS(cfg, nil)
}

// NewRepositoryWithFS allows passing a custom FS (like MemoryFS for tests)
func NewRepositoryWithFS(cfg *config.RepoConfig, targetFS fs.FS) (*Repository, error) {
	if targetFS == nil {
		targetFS = &fs.OSFS{}
	}

	// Create Meta
	meta, err := meta.NewMeta(cfg, targetFS)
	if err != nil {
		return nil, err
	}

	// Create BlockContext
	block := block.NewBlockContext(cfg.BlocksDir(), targetFS)

	// Create FileContext
	file := file.NewFileContext(cfg.WorkingTreeDir, cfg.RepoDir, block, targetFS)

	// Create file scanner
	scanner := scan.NewScanner(cfg.WorkingTreeDir, meta, targetFS)

	// Create SnapshotContext
	snapshot := snapshot.NewSnapshotContext(cfg.SnapshotsDir(), file, block, scanner, targetFS)

	// Ensure store layout
	if !isStoreExists(cfg, targetFS) {
		if err := createStoreStructure(cfg, targetFS); err != nil {
			return nil, err
		}
	}

	return &Repository{
		Config:   cfg,
		Meta:     meta,
		Block:    block,
		File:     file,
		Snapshot: snapshot,
		FS:       targetFS,
	}, nil
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

func (r *Repository) writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)

	tmpFile, tmpPath, err := r.FS.CreateTempFile(dir, "tmp-*.json")
	if err != nil {
		return err
	}
	defer r.FS.Remove(tmpPath) // ensure cleanup on error

	// write JSON
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// atomically rename
	return r.FS.Rename(tmpPath, path)
}

func (r *Repository) readJSON(path string, v interface{}) error {
	data, err := r.FS.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
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
