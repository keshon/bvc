package file_test

import (
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
	"path/filepath"
	"testing"
)

func TestBuildEntryAndEntries(t *testing.T) {
	tmpDir := t.TempDir()
	fs := newMockFS()
	blocks := newMockBlock()

	// pretend this is the repo’s working tree
	fc := &file.FileContext{
		Root:     tmpDir,
		RepoRoot: tmpDir,
		FS:       fs,
		Blocks:   blocks,
	}

	// create a file inside the working tree
	f := filepath.Join(tmpDir, "foo.txt")
	fs.WriteFile(f, []byte("hello"), 0o644)

	// test BuildEntry
	entry, err := fc.BuildEntry(f)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Path != "foo.txt" { // relative to Root
		t.Errorf("expected relative path 'foo.txt', got %q", entry.Path)
	}

	// test BuildEntries
	entries, err := fc.BuildEntries([]string{f}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !entries[0].Equal(&entry) {
		t.Errorf("BuildEntries mismatch: got %+v", entries)
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	fs := newMockFS()
	fc := &file.FileContext{FS: fs, Root: tmpDir}

	existing := filepath.Join(tmpDir, "exists.txt")
	fs.WriteFile(existing, []byte("ok"), 0o644)

	if !fc.Exists(existing) {
		t.Error("Exists should return true")
	}
	if fc.Exists(filepath.Join(tmpDir, "missing.txt")) {
		t.Error("Exists should return false")
	}
}

func TestBuildEntryErrors(t *testing.T) {
	tmpDir := t.TempDir()
	fs := newMockFS()
	blocks := newMockBlock()
	blocks.files = nil // simulate block read missing
	fc := &file.FileContext{
		Root:   tmpDir,
		FS:     fs,
		Blocks: blocks,
	}

	// missing file
	_, err := fc.BuildEntry("/missing.txt")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}

	// restore with missing block
	entry := file.Entry{
		Path:   "/foo.txt",
		Blocks: []block.BlockRef{{Hash: "nonexistent", Size: 3}},
	}
	err = fc.RestoreFilesToWorkingTree([]file.Entry{entry}, "test")
	if err == nil {
		t.Error("expected error when block missing, got nil")
	}
}

func TestBuildEntryNilBlocks(t *testing.T) {
	fc := &file.FileContext{FS: newMockFS(), Root: "/"}
	_, err := fc.BuildEntry("foo.txt")
	if err == nil {
		t.Error("expected error when Blocks is nil")
	}
}

func TestWriteNilBlocks(t *testing.T) {
	fc := &file.FileContext{FS: newMockFS(), Root: "/"}
	entry := file.Entry{Path: "x.txt"}
	if err := fc.Write(entry); err == nil {
		t.Error("expected error when writing with nil Blocks")
	}
}

func TestEntryEqualEdgeCases(t *testing.T) {
	var e1, e2 file.Entry
	if !e1.Equal(&e2) {
		t.Error("empty entries should be equal")
	}
	if e1.Equal(nil) {
		t.Error("entry.Equal(nil) should be false")
	}
	e1.Blocks = []block.BlockRef{{Hash: "a"}}
	e2.Blocks = []block.BlockRef{{Hash: "b"}}
	if e1.Equal(&e2) {
		t.Error("entries with different block hashes should not be equal")
	}
}
