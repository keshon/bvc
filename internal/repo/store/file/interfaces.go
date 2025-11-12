package file

import (
	"app/internal/repo/store/block"
	"os"
)

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

// BlockStore abstracts block operations.
type BlockStore interface {
	SplitFile(path string) ([]block.BlockRef, error)
	Write(path string, blocks []block.BlockRef) error
	Read(hash string) ([]byte, error)
}
