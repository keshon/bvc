package file_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"app/internal/fsio"
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
)

// helpers
func makeTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

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

func makeBlockContext(t *testing.T, dir string) *block.BlockContext {
	t.Helper()
	return &block.BlockContext{Root: dir}
}

// --- BuildEntry / Write / Exists --- //
func TestBuildEntryWriteExists(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	fm := &file.FileContext{
		Root:   dir,
		Blocks: makeBlockContext(t, dir),
	}

	content := []byte("hello world")
	filePath := filepath.Join(dir, "file1.txt")
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	entry, err := fm.BuildEntry(filePath)
	if err != nil {
		t.Fatalf("BuildEntry failed: %v", err)
	}

	if err := fm.Write(entry); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if !fm.Exists(filePath) {
		t.Errorf("Exists returned false for existing file")
	}
}

// --- BuildEntries --- //
func TestBuildEntries(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	fm := &file.FileContext{
		Root:   dir,
		Blocks: makeBlockContext(t, dir),
	}

	files := []string{}
	for i := 0; i < 3; i++ {
		path := filepath.Join(dir, "file"+string(rune('a'+i))+".txt")
		os.WriteFile(path, []byte("data"), 0o644)
		files = append(files, path)
	}

	entries, err := fm.BuildEntries(files)
	if err != nil {
		t.Fatalf("BuildEntries failed: %v", err)
	}

	if len(entries) != len(files) {
		t.Errorf("expected %d entries, got %d", len(files), len(entries))
	}
}

// --- SaveIndex / LoadIndex / ClearIndex --- //
func TestStageAndLoadIndex(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	fm := &file.FileContext{Root: dir}

	entry := file.Entry{Path: "a.txt", Blocks: nil}
	if err := fm.SaveIndex([]file.Entry{entry}); err != nil {
		t.Fatalf("SaveIndex failed: %v", err)
	}

	loaded, err := fm.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Path != "a.txt" {
		t.Errorf("loaded entries mismatch: %+v", loaded)
	}

	if err := fm.ClearIndex(); err != nil {
		t.Fatalf("ClearIndex failed: %v", err)
	}
	loaded, _ = fm.LoadIndex()
	if loaded != nil {
		t.Errorf("expected nil after ClearIndex, got %+v", loaded)
	}
}

// --- ListAll --- //
func TestListAll(t *testing.T) {
	dir := t.TempDir()
	fm := &file.FileContext{Root: dir}

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0o755)
	os.WriteFile(filepath.Join(dir, "subdir/b.txt"), []byte("y"), 0o644)

	paths, _, err := fm.ScanFilesInWorkingTree()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("expected 2 files, got %d", len(paths))
	}
}

func TestFileContext_ErrorBranches(t *testing.T) {
	tmp := makeTempDir(t)
	fm := &file.FileContext{Root: tmp}

	// 1. BuildEntry/Write with nil BlockContext
	if _, err := fm.BuildEntry("a.txt"); err == nil {
		t.Error("expected error for BuildEntry with nil Blocks")
	}
	if err := fm.Write(file.Entry{Path: "x"}); err == nil {
		t.Error("expected error for Write with nil Blocks")
	}

	// 2. Exists on missing file
	if fm.Exists(filepath.Join(tmp, "nofile.txt")) {
		t.Error("expected Exists to return false")
	}

	// 3. SaveIndex marshal error (use circular data to break json.Marshal)
	// Inject a self-referential value to cause json.Marshal to fail
	type loop struct {
		Next *loop
	}
	bad := []loop{{}}
	bad[0].Next = &bad[0]
	// Temporarily replace fsio.WriteFile to simulate JSON encoding failure
	oldWrite := fsio.WriteFile
	fsio.WriteFile = func(string, []byte, os.FileMode) error {
		return errors.New("fake write error")
	}
	defer func() { fsio.WriteFile = oldWrite }()

	if err := fm.SaveIndex([]file.Entry{{Path: "a"}}); err == nil {
		t.Error("expected SaveIndex write failure")
	}

	// 4. SaveIndex mkdir error — simulate by patching fsio.MkdirAll
	oldMkdir := fsio.MkdirAll
	fsio.MkdirAll = func(string, os.FileMode) error {
		return errors.New("fake mkdir error")
	}
	defer func() { fsio.MkdirAll = oldMkdir }()
	if err := fm.SaveIndex([]file.Entry{}); err == nil {
		t.Error("expected mkdir failure")
	}

	// 5. ClearIndex with missing file
	if err := fm.ClearIndex(); err != nil {
		t.Error("ClearIndex should not error on missing file")
	}

	// 6. LoadIndex missing index.json
	if entries, err := fm.LoadIndex(); err != nil || entries != nil {
		t.Error("expected nil,nil for missing index.json")
	}

	// 7. LoadIndex bad JSON
	idx := filepath.Join(tmp, "index.json")
	os.WriteFile(idx, []byte("{ bad json"), 0o644)
	if _, err := fm.LoadIndex(); err == nil {
		t.Error("expected unmarshal error")
	}

	// 8. Equal edge cases
	var e1, e2 file.Entry
	if !e1.Equal(&e2) {
		t.Error("expected empty Equal true")
	}
	if e1.Equal(nil) {
		t.Error("nil mismatch should be false")
	}
	e1.Blocks = []block.BlockRef{{Hash: "a"}}
	e2.Blocks = []block.BlockRef{{Hash: "b"}}
	if e1.Equal(&e2) {
		t.Error("different hash should be false")
	}
}

func TestCreateAllAndChangedEntriesErrors(t *testing.T) {
	tmp := makeTempDir(t)
	fm := &file.FileContext{
		Root:   tmp,
		Blocks: makeBlockContext(t, tmp),
	}

	// BuildAllEntries with no files — should not panic
	_, _ = fm.BuildAllEntries()

	// BuildChangedEntries with empty index
	if _, err := fm.BuildChangedEntries(); err != nil {
		t.Errorf("unexpected error on empty index: %v", err)
	}
}

// Simple integration test for FileContext + BlockContext SplitFile
func TestSplitFileIntegration(t *testing.T) {
	dir := t.TempDir()
	bm := &block.BlockContext{Root: dir}
	fm := &file.FileContext{Root: dir, Blocks: bm}

	content := "abcdefghijklmnopqrstuvwxyz" // small file, simple deterministic content
	testFile := makeSplitTestFile(t, content)
	defer os.Remove(testFile)

	// Create entry
	entry, err := fm.BuildEntry(testFile)
	if err != nil {
		t.Fatalf("BuildEntry failed: %v", err)
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
