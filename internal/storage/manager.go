package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"app/internal/config"
	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
)

// Manager is the high-level storage abstraction that unifies all subsystems.
type Manager struct {
	Root      string
	Objects   string
	Commits   string
	Branches  string
	Filesets  string
	Blocks    *block.BlockManager
	Files     *file.FileManager
	Snapshots *snapshot.SnapshotManager
}

// InitAt sets up the directory structure under root (usually .bvc/).
func InitAt(root string) (*Manager, error) {
	// Default root
	if root == "" {
		root = config.DetectRepoRoot()
	}

	root = filepath.Clean(root)

	// Ensure required structure exists
	dirs := []string{
		filepath.Join(root, config.CommitsDir),
		filepath.Join(root, config.FilesetsDir),
		filepath.Join(root, config.BranchesDir),
		filepath.Join(root, config.ObjectsDir),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("init storage dir %q: %w", d, err)
		}
	}

	return NewManager(root), nil
}

// NewManager constructs a new storage manager (called by InitAt).
func NewManager(root string) *Manager {
	root = filepath.Clean(root)

	m := &Manager{
		Root:     root,
		Objects:  filepath.Join(root, config.ObjectsDir),
		Commits:  filepath.Join(root, config.CommitsDir),
		Branches: filepath.Join(root, config.BranchesDir),
		Filesets: filepath.Join(root, config.FilesetsDir),
	}

	m.Blocks = &block.BlockManager{Root: m.Objects}
	m.Files = &file.FileManager{Root: root, Blocks: m.Blocks}
	m.Snapshots = &snapshot.SnapshotManager{
		Root:   m.Filesets,
		Files:  m.Files,
		Blocks: m.Blocks,
	}

	return m
}
