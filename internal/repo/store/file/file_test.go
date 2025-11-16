package file_test

import (
	"app/internal/fs"
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

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

// Helper to create FileContext with in-memory FS.
func newTestBC(t *testing.T) (*block.BlockContext, string) {
	t.Helper()
	tmpDir := t.TempDir()
	fs := fs.NewMemoryFS()
	err := fs.MkdirAll(filepath.Join(tmpDir, "blocks"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	blockCtx := block.NewBlockContext(filepath.Join(tmpDir, "blocks"), fs)

	return blockCtx, tmpDir
}

func newTestFC(t *testing.T) (*file.FileContext, string) {
	t.Helper()
	tmpDir := t.TempDir()
	fs := fs.NewMemoryFS()
	blockCtx, _ := newTestBC(t)

	tmpRepoRoot := filepath.Join(t.TempDir(), ".bvc")
	err := fs.MkdirAll(tmpRepoRoot, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	fc := file.NewFileContext(tmpDir, tmpRepoRoot, blockCtx, fs)
	return fc, tmpDir
}
