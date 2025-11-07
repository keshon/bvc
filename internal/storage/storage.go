package storage

import (
	"fmt"

	"app/internal/config"
	"app/internal/fsio"
	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
)

// Manager is the high-level storage abstraction that unifies all subsystems.
type Manager struct {
	Config    *config.RepoConfig
	Blocks    *block.BlockManager
	Files     *file.FileManager
	Snapshots *snapshot.SnapshotManager
}

// InitAt sets up the directory structure under root (usually .bvc/).
func InitAt(root string) (*Manager, error) {
	cfg := config.NewRepoConfig(root)

	dirs := []string{
		cfg.CommitsDir(),
		cfg.FilesetsDir(),
		cfg.BranchesDir(),
		cfg.ObjectsDir(),
	}

	for _, d := range dirs {
		if err := fsio.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("init storage dir %q: %w", d, err)
		}
	}

	return NewManager(cfg), nil
}

// OpenAt opens a storage manager for an existing repo.
func OpenAt(root string) (*Manager, error) {
	cfg := config.NewRepoConfig(root)
	return NewManager(cfg), nil
}

// NewManager constructs a storage manager (internal).
func NewManager(cfg *config.RepoConfig) *Manager {
	m := &Manager{Config: cfg}

	m.Blocks = &block.BlockManager{Root: cfg.ObjectsDir()}
	m.Files = &file.FileManager{Root: cfg.Root, Blocks: m.Blocks}
	m.Snapshots = &snapshot.SnapshotManager{
		Root:   cfg.FilesetsDir(),
		Files:  m.Files,
		Blocks: m.Blocks,
	}

	return m
}
