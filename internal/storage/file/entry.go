package file

import (
	"app/internal/fsio"
	"app/internal/storage/block"
	"fmt"
	"path/filepath"
)

// Entry represents a tracked file and its content blocks.
type Entry struct {
	Path   string
	Blocks []block.BlockRef
}

// Equal compares two entries by their block lists.
func (e *Entry) Equal(other *Entry) bool {
	if e == nil && other == nil {
		return true
	}
	if e == nil || other == nil {
		return false
	}
	if len(e.Blocks) != len(other.Blocks) {
		return false
	}
	for i := range e.Blocks {
		if e.Blocks[i].Hash != other.Blocks[i].Hash ||
			e.Blocks[i].Size != other.Blocks[i].Size {
			return false
		}
	}
	return true
}

// FileManager wraps file-level operations that depend on BlockManager.
type FileManager struct {
	Root   string
	Blocks *block.BlockManager
}

// CreateEntry splits a file into block references (content-defined).
func (fm *FileManager) CreateEntry(path string) (Entry, error) {
	if fm.Blocks == nil {
		return Entry{}, fmt.Errorf("no BlockManager attached")
	}
	blocks, err := fm.Blocks.SplitFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("split %q: %w", path, err)
	}
	return Entry{Path: path, Blocks: blocks}, nil
}

// Write stores all blocks of an entry into storage.
func (fm *FileManager) Write(e Entry) error {
	if fm.Blocks == nil {
		return fmt.Errorf("no BlockManager attached")
	}
	return fm.Blocks.Write(e.Path, e.Blocks)
}

// Exists checks whether a given path exists in the working tree.
func (fm *FileManager) Exists(path string) bool {
	_, err := fsio.StatFile(filepath.Clean(path))
	return err == nil
}
