package file

import (
	"os"
)

// OSFS is a production implementation of FS using the standard library.
type OSFS struct{}

func (r *OSFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (r *OSFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (r *OSFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (r *OSFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r *OSFS) Remove(path string) error {
	return os.Remove(path)
}

func (r *OSFS) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (r *OSFS) CreateTempFile(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func (r *OSFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
