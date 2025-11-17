package file_test

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/store/block"
	"github.com/keshon/bvc/internal/repo/store/file"
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

func (b *mockBlock) BlocksDir() string { return "" }

// Helper to create FileContext with in-memory FS.
func newTestFC(t *testing.T) (*file.FileContext, string) {
	t.Helper()
	tmpDir := t.TempDir()

	// ONE shared MemoryFS
	mem := fs.NewMemoryFS()

	// create blocks dir inside repo
	repoRoot := filepath.Join(tmpDir, ".bvc")
	if err := mem.MkdirAll(filepath.Join(repoRoot, "blocks"), 0o755); err != nil {
		t.Fatal(err)
	}

	// create BlockContext using the SAME FS
	blockCtx := block.NewBlockContext(filepath.Join(repoRoot, "blocks"), mem)

	// FileContext using the same FS + same BlockCtx
	fc := file.NewFileContext(tmpDir, repoRoot, blockCtx, mem)

	return fc, tmpDir
}
