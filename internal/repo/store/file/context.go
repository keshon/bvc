package file

import (
	"app/internal/repo/store/block"
	"os"
)

// Entry represents a tracked file and its content blocks.
type Entry struct {
	Path   string
	Blocks []block.BlockRef
}

// Equal compares two entries by their blocks.
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
		if e.Blocks[i].Hash != other.Blocks[i].Hash || e.Blocks[i].Size != other.Blocks[i].Size {
			return false
		}
	}
	return true
}

// FS abstracts filesystem operations.
type FS interface {
	Stat(path string) (os.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	Rename(oldPath, newPath string) error
	CreateTempFile(dir, pattern string) (*os.File, error)
	IsNotExist(err error) bool
}

// BlockContext abstracts block operations.
type BlockContext interface {
	SplitFile(path string) ([]block.BlockRef, error)
	Write(path string, blocks []block.BlockRef) error
	Read(hash string) ([]byte, error)
}

// FileContext manages file-level operations (staging, restore, scan) with abstracted dependencies.
type FileContext struct {
	Root     string
	RepoRoot string
	FS       FS
	Blocks   BlockContext
}
