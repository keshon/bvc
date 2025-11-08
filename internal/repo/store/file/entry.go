package file

import (
	"app/internal/fsio"
	"app/internal/repo/store/block"

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

// FileContext wraps file-level operations that depend on BlockContext.
type FileContext struct {
	Root     string
	RepoRoot string
	Blocks   *block.BlockContext
}

// CreateEntry splits a file into block references (content-defined).
func (fc *FileContext) CreateEntry(path string) (Entry, error) {
	if fc.Blocks == nil {
		return Entry{}, fmt.Errorf("no BlockContext attached")
	}
	blocks, err := fc.Blocks.SplitFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("split %q: %w", path, err)
	}
	return Entry{Path: path, Blocks: blocks}, nil
}

// Write stores all blocks of an entry into store.
func (fc *FileContext) Write(e Entry) error {
	if fc.Blocks == nil {
		return fmt.Errorf("no BlockContext attached")
	}
	return fc.Blocks.Write(e.Path, e.Blocks)
}

// Exists checks whether a given path exists in the working tree.
func (fc *FileContext) Exists(path string) bool {
	_, err := fsio.StatFile(filepath.Clean(path))
	return err == nil
}
