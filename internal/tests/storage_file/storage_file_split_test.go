package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"app/internal/storage/block"
	"app/internal/storage/file"
)

func makeSplitTestFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "bvc-split-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

// Simple integration test for FileManager + BlockManager SplitFile
func TestSplitFileIntegration(t *testing.T) {
	dir := t.TempDir()
	bm := &block.BlockManager{Root: dir}
	fm := &file.FileManager{Root: dir, Blocks: bm}

	content := "abcdefghijklmnopqrstuvwxyz" // small file, simple deterministic content
	testFile := makeSplitTestFile(t, content)
	defer os.Remove(testFile)

	// Create entry
	entry, err := fm.CreateEntry(testFile)
	if err != nil {
		t.Fatalf("CreateEntry failed: %v", err)
	}

	if len(entry.Blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}

	// Write blocks
	if err := fm.Write(entry); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify blocks exist
	for _, b := range entry.Blocks {
		path := filepath.Join(bm.Root, b.Hash+".bin")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("block file %q missing: %v", path, err)
		}
		if info.Size() != b.Size {
			t.Errorf("block size mismatch: got %d, want %d", info.Size(), b.Size)
		}
	}
}
