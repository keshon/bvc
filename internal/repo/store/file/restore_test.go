package file_test

import (
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreFiles(t *testing.T) {
	tmp := t.TempDir()
	fs := newMockFS()
	blocks := newMockBlock()
	fc := &file.FileContext{FS: fs, Blocks: blocks, Root: tmp}

	target := filepath.Join(tmp, "foo.txt")
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

func TestSplitFileIntegration(t *testing.T) {
	dir := t.TempDir()
	bm := &block.BlockContext{Root: dir} // real FS-backed block storage
	fc := &file.FileContext{Root: dir, Blocks: bm}

	content := "abcdefghijklmnopqrstuvwxyz"
	tmpFilePath := filepath.Join(dir, "bvc-split-test.txt")
	if err := os.WriteFile(tmpFilePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	entry, err := fc.BuildEntry(tmpFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entry.Blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}

	if err := fc.Write(entry); err != nil {
		t.Fatal(err)
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
