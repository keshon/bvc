package fs

import "os"

// OSFS is a production implementation of FS using the standard library.
type OSFS struct{}

func NewOSFS() *OSFS {
	return &OSFS{}
}

func (r *OSFS) Open(path string) (*os.File, error) {
	return open(path)
}

func (r *OSFS) Stat(path string) (os.FileInfo, error) {
	return stat(path)
}

func (r *OSFS) ReadFile(path string) ([]byte, error) {
	return readFile(path)
}

func (r *OSFS) ReadDir(path string) ([]os.DirEntry, error) {
	return readDir(path)
}

func (r *OSFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return writeFile(path, data, perm)
}

func (r *OSFS) MkdirAll(path string, perm os.FileMode) error {
	return mkdirAll(path, perm)
}

func (r *OSFS) Remove(path string) error {
	return remove(path)
}

func (r *OSFS) Rename(oldPath, newPath string) error {
	return rename(oldPath, newPath)
}

func (r *OSFS) CreateTempFile(dir, pattern string) (*os.File, error) {
	return createTemp(dir, pattern)
}

func (r *OSFS) IsNotExist(err error) bool {
	return isNotExist(err)
}

func (r *OSFS) IsDir(path string) bool {
	return IsDir(path)
}

func (r *OSFS) Exists(path string) bool {
	return exists(path)
}
