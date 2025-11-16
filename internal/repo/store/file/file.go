package file

import (
	"app/internal/fs"
	"app/internal/repo/store/block"
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

// BlockContext abstracts block operations.
type BlockContext interface {
	SplitFile(path string) ([]block.BlockRef, error)
	Write(path string, blocks []block.BlockRef) error
	Read(hash string) ([]byte, error)
}

// FileContext manages file-level operations (staging, restore, scan) with abstracted dependencies.
type FileContext struct {
	WorkingTreeDir string
	RepoDir        string
	BlockCtx       BlockContext
	FS             fs.FS
}

// NewFileContext creates a new FileContext.
func NewFileContext(workingTreeDir, repoDir string, blocks BlockContext, fs fs.FS) *FileContext {
	return &FileContext{WorkingTreeDir: workingTreeDir, RepoDir: repoDir, BlockCtx: blocks, FS: fs}
}
