package fs

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
)

// CompressedFS wraps another FS and compresses all file writes.
type CompressedFS struct {
	underlying FS
}

func NewCompressedFS(base FS) *CompressedFS {
	return &CompressedFS{underlying: base}
}

func (c *CompressedFS) Open(path string) (io.ReadSeekCloser, error) {
	rc, err := c.underlying.Open(path)
	if err != nil {
		return nil, err
	}

	// read all compressed data and decompress
	data, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewReader(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	return &memReadSeekCloser{Reader: bytes.NewReader(decompressed)}, nil
}

func (c *CompressedFS) ReadFile(path string) ([]byte, error) {
	rc, err := c.Open(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func (c *CompressedFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return err
	}
	gz.Close()
	return c.underlying.WriteFile(path, buf.Bytes(), perm)
}

// Pass-through for other operations
func (c *CompressedFS) MkdirAll(path string, perm os.FileMode) error {
	return c.underlying.MkdirAll(path, perm)
}
func (c *CompressedFS) Remove(path string) error { return c.underlying.Remove(path) }
func (c *CompressedFS) Rename(oldPath, newPath string) error {
	return c.underlying.Rename(oldPath, newPath)
}
func (c *CompressedFS) Stat(path string) (os.FileInfo, error)      { return c.underlying.Stat(path) }
func (c *CompressedFS) ReadDir(path string) ([]os.DirEntry, error) { return c.underlying.ReadDir(path) }
func (c *CompressedFS) CreateTempFile(dir, pattern string) (io.WriteCloser, string, error) {
	return c.underlying.CreateTempFile(dir, pattern)
}
func (c *CompressedFS) IsNotExist(err error) bool { return c.underlying.IsNotExist(err) }
func (c *CompressedFS) IsDir(path string) bool    { return c.underlying.IsDir(path) }
func (c *CompressedFS) Exists(path string) bool   { return c.underlying.Exists(path) }
