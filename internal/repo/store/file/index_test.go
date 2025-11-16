package file_test

import (
	"path/filepath"
	"testing"

	"github.com/keshon/bvc/internal/repo/store/file"
)

func TestIndexCRUD(t *testing.T) {
	fc, _ := newTestFC(t)

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
	fc, _ := newTestFC(t)

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
	fc, _ := newTestFC(t)

	// missing index.json
	entries, err := fc.LoadIndex()
	if err != nil || entries != nil {
		t.Error("expected nil,nil for missing index.json")
	}

	// invalid JSON
	idx := filepath.Join(fc.RepoDir, "index.json")
	err = fc.FS.WriteFile(idx, []byte("{ bad json"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fc.LoadIndex(); err == nil {
		t.Error("expected unmarshal error for bad JSON")
	}
}

func TestClearIndexMissingFile(t *testing.T) {
	fc, _ := newTestFC(t)

	// should not fail even if index.json doesn't exist
	if err := fc.ClearIndex(); err != nil {
		t.Error("ClearIndex should succeed on missing file")
	}
}
