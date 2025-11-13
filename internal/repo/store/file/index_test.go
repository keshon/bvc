package file_test

import (
	"app/internal/repo/store/file"
	"path/filepath"
	"testing"
)

func TestIndexCRUD(t *testing.T) {
	tmp := t.TempDir()
	fs := newMockFS()
	fc := &file.FileContext{RepoRoot: tmp, FS: fs}

	entries := []file.Entry{{Path: "a.txt"}}
	if err := fc.SaveIndexReplace(entries); err != nil {
		t.Fatal(err)
	}

	loaded, err := fc.LoadIndex()
	if err != nil || len(loaded) != 1 {
		t.Fatalf("LoadIndex failed: %v", err)
	}

	if err := fc.ClearIndex(); err != nil {
		t.Fatal(err)
	}
	loaded, _ = fc.LoadIndex()
	if len(loaded) != 0 {
		t.Error("index not cleared")
	}
}

func TestSaveIndexMerge(t *testing.T) {
	tmp := t.TempDir()
	fs := newMockFS()
	fc := &file.FileContext{RepoRoot: tmp, FS: fs}

	initial := []file.Entry{{Path: "a.txt"}}
	if err := fc.SaveIndexReplace(initial); err != nil {
		t.Fatal(err)
	}

	newEntries := []file.Entry{{Path: "b.txt"}}
	if err := fc.SaveIndexMerge(newEntries); err != nil {
		t.Fatal(err)
	}

	loaded, err := fc.LoadIndex()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 entries after merge, got %d", len(loaded))
	}
}

func TestLoadIndexMissingAndInvalid(t *testing.T) {
	tmp := t.TempDir()
	fc := &file.FileContext{FS: newMockFS(), RepoRoot: tmp}

	// missing index.json
	entries, err := fc.LoadIndex()
	if err != nil || entries != nil {
		t.Error("expected nil,nil for missing index.json")
	}

	// invalid JSON
	idx := filepath.Join(tmp, "index.json")
	fs := newMockFS()
	fs.WriteFile(idx, []byte("{ bad json"), 0o644) // write into mockFS
	fc = &file.FileContext{FS: fs, RepoRoot: tmp}

	if _, err := fc.LoadIndex(); err == nil {
		t.Error("expected unmarshal error for bad JSON")
	}
}

func TestClearIndexMissingFile(t *testing.T) {
	tmp := t.TempDir()
	fc := &file.FileContext{FS: newMockFS(), RepoRoot: tmp}
	// should not fail even if index.json doesn't exist
	if err := fc.ClearIndex(); err != nil {
		t.Error("ClearIndex should succeed on missing file")
	}
}
