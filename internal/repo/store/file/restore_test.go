package file_test

import (
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
	"path/filepath"
	"testing"
)

func TestRestoreFiles(t *testing.T) {
	tmpDir := t.TempDir()
	fs := newMockFS()
	blocks := newMockBlock()
	fc := &file.FileContext{FS: fs, Blocks: blocks, Root: tmpDir}

	target := filepath.Join(tmpDir, "foo.txt")
	entry := file.Entry{
		Path:   target,
		Blocks: []block.BlockRef{{Hash: "hash1", Size: 3}},
	}

	// ensure mockBlock.Read will return desired payload for hash1 (we used "data" in mock)
	blocks.files["hash1"] = []byte("data")

	if err := fc.RestoreFilesToWorkingTree([]file.Entry{entry}, "test"); err != nil {
		t.Fatal(err)
	}

	data, _ := fs.ReadFile(target)
	if string(data) != "data" {
		t.Errorf("unexpected file content %q", data)
	}
}
