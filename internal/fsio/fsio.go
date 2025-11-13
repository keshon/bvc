package fsio

import (
	"os"
)

// FSIO is a production implementation of FS using the standard library.
type FSIO struct{}

func (r *FSIO) Stat(path string) (os.FileInfo, error) {
	return Stat(path)
}

func (r *FSIO) ReadFile(path string) ([]byte, error) {
	return ReadFile(path)
}

func (r *FSIO) WriteFile(path string, data []byte, perm os.FileMode) error {
	return WriteFile(path, data, perm)
}

func (r *FSIO) MkdirAll(path string, perm os.FileMode) error {
	return MkdirAll(path, perm)
}

func (r *FSIO) Remove(path string) error {
	return Remove(path)
}

func (r *FSIO) Rename(oldPath, newPath string) error {
	return Rename(oldPath, newPath)
}

func (r *FSIO) CreateTempFile(dir, pattern string) (*os.File, error) {
	return CreateTemp(dir, pattern)
}

func (r *FSIO) IsNotExist(err error) bool {
	return IsNotExist(err)
}
