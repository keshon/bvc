package fs

import (
	"io"
	"os"
)

// FS abstracts filesystem operations.
type FS interface {
	Open(path string) (io.ReadSeekCloser, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	Rename(oldPath, newPath string) error
	Stat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.DirEntry, error)
	CreateTempFile(dir, pattern string) (io.WriteCloser, string, error)
	IsNotExist(err error) bool
	Exists(path string) bool
	IsDir(path string) bool
}
