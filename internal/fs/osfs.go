package fs

import (
	"io"
	"os"
)

// OSFS is a production FS implementation using the standard library.
type OSFS struct{}

func NewOSFS() *OSFS {
	return &OSFS{}
}

func (fsys *OSFS) Open(path string) (io.ReadSeekCloser, error) {
	return open(path)
}

func (fsys *OSFS) Stat(path string) (os.FileInfo, error) {
	return stat(path)
}

func (fsys *OSFS) ReadFile(path string) ([]byte, error) {
	return readFile(path)
}

func (fsys *OSFS) ReadDir(path string) ([]os.DirEntry, error) {
	return readDir(path)
}

func (fsys *OSFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return writeFile(path, data, perm)
}

func (fsys *OSFS) MkdirAll(path string, perm os.FileMode) error {
	return mkdirAll(path, perm)
}

func (fsys *OSFS) Remove(path string) error {
	return remove(path)
}

func (fsys *OSFS) Rename(oldPath, newPath string) error {
	return rename(oldPath, newPath)
}

func (fsys *OSFS) CreateTempFile(dir, pattern string) (io.WriteCloser, string, error) {
	return createTemp(dir, pattern)
}

func (fsys *OSFS) IsNotExist(err error) bool {
	return isNotExist(err)
}

func (fsys *OSFS) IsDir(path string) bool {
	return IsDir(path)
}

func (fsys *OSFS) Exists(path string) bool {
	return exists(path)
}
