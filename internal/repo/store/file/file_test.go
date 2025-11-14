package file_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"app/internal/repo/store/block"
)

// mockFS is an in-memory FS for testing
type mockFS struct {
	files map[string][]byte
	mu    sync.Mutex
}

func newMockFS() *mockFS {
	return &mockFS{files: make(map[string][]byte)}
}

func (m *mockFS) Stat(path string) (os.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.files[path]; ok {
		return &mockFileInfo{name: filepath.Base(path), size: int64(len(m.files[path]))}, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFS) Open(path string) (*os.File, error) {
	return os.Open(path)
}

func (m *mockFS) ReadFile(path string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// try in-memory map first
	if data, ok := m.files[path]; ok {
		return append([]byte(nil), data...), nil
	}

	// fallback to real FS (for temp files created by CreateTempFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (m *mockFS) ReadDir(path string) ([]os.DirEntry, error) {
	return nil, nil
}

func (m *mockFS) CreateTempFile(dir, pattern string) (*os.File, error) {
	// use the real os.CreateTemp to obtain a writable *os.File
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, err
	}
	// ensure the real file is removed when Close+Rename happen — mockFS.Rename will import it
	return f, nil
}

func (m *mockFS) Rename(oldPath, newPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if file exists in memory map, just move it
	if data, ok := m.files[oldPath]; ok {
		delete(m.files, oldPath)
		m.files[newPath] = data
		return nil
	}

	// otherwise try to read from real filesystem (temp file created by CreateTempFile)
	data, err := os.ReadFile(oldPath)
	if err != nil {
		// not present on disk either
		return os.ErrNotExist
	}

	// move into in-memory map and delete real file
	m.files[newPath] = append([]byte(nil), data...)
	_ = os.Remove(oldPath)
	return nil
}

func (m *mockFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = append([]byte(nil), data...)
	return nil
}

func (m *mockFS) MkdirAll(path string, perm os.FileMode) error { return nil }
func (m *mockFS) Remove(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, path)
	return nil
}

func (m *mockFS) IsNotExist(err error) bool { return errors.Is(err, os.ErrNotExist) }

func (m *mockFS) IsDir(path string) bool { return false }

func (m *mockFS) Exists(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.files[path]
	return ok
}

type mockBlock struct {
	files map[string][]byte
	mu    sync.Mutex
}

func newMockBlock() *mockBlock {
	return &mockBlock{files: make(map[string][]byte)}
}

func (b *mockBlock) SplitFile(path string) ([]block.BlockRef, error) {
	return []block.BlockRef{{Hash: path + "-hash", Size: 123}}, nil
}
func (b *mockBlock) Write(path string, blocks []block.BlockRef) error { return nil }
func (b *mockBlock) Read(hash string) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.files == nil {
		return nil, fmt.Errorf("block storage not initialized")
	}
	data, ok := b.files[hash]
	if !ok {
		return nil, fmt.Errorf("block %s not found", hash)
	}
	return append([]byte(nil), data...), nil
}

type mockFileInfo struct {
	name string
	size int64
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0o644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }
