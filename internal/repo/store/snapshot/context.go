package snapshot

import (
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
)

// SnapshotContext handles higher-level operations (filesets, commits)
type SnapshotContext struct {
	Root   string
	Files  *file.FileContext
	Blocks *block.BlockContext
}

// Fileset represents a snapshot of tracked files and their block mappings.
type Fileset struct {
	ID    string       `json:"id"`
	Files []file.Entry `json:"files"`
}
