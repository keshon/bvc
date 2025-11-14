package fs

import "os"

// FS abstracts filesystem operations.
type FS interface {
	Open(path string) (*os.File, error)
	Stat(path string) (os.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]os.DirEntry, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	Rename(oldPath, newPath string) error
	CreateTempFile(dir, pattern string) (*os.File, error)
	IsNotExist(err error) bool
	IsDir(path string) bool
	Exists(path string) bool
}
