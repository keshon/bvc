package file_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"app/internal/fsio"
	"app/internal/storage/block"
	"app/internal/storage/file"
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

func makeBlockManager(t *testing.T, dir string) *block.BlockManager {
	t.Helper()
	return &block.BlockManager{Root: dir}
}

// --- CreateEntry / Write / Exists --- //
func TestCreateEntryWriteExists(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	fm := &file.FileManager{
		Root:   dir,
		Blocks: makeBlockManager(t, dir),
	}

	content := []byte("hello world")
	filePath := filepath.Join(dir, "file1.txt")
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	entry, err := fm.CreateEntry(filePath)
	if err != nil {
		t.Fatalf("CreateEntry failed: %v", err)
	}

	if err := fm.Write(entry); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if !fm.Exists(filePath) {
		t.Errorf("Exists returned false for existing file")
	}
}

// --- CreateEntries --- //
func TestCreateEntries(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	fm := &file.FileManager{
		Root:   dir,
		Blocks: makeBlockManager(t, dir),
	}

	files := []string{}
	for i := 0; i < 3; i++ {
		path := filepath.Join(dir, "file"+string(rune('a'+i))+".txt")
		os.WriteFile(path, []byte("data"), 0o644)
		files = append(files, path)
	}

	entries, err := fm.CreateEntries(files)
	if err != nil {
		t.Fatalf("CreateEntries failed: %v", err)
	}

	if len(entries) != len(files) {
		t.Errorf("expected %d entries, got %d", len(files), len(entries))
	}
}

// --- StageFiles / GetIndexFiles / ClearIndex --- //
func TestStageAndLoadIndex(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	fm := &file.FileManager{Root: dir}

	entry := file.Entry{Path: "a.txt", Blocks: nil}
	if err := fm.StageFiles([]file.Entry{entry}); err != nil {
		t.Fatalf("StageFiles failed: %v", err)
	}

	loaded, err := fm.GetIndexFiles()
	if err != nil {
		t.Fatalf("GetIndexFiles failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Path != "a.txt" {
		t.Errorf("loaded entries mismatch: %+v", loaded)
	}

	if err := fm.ClearIndex(); err != nil {
		t.Fatalf("ClearIndex failed: %v", err)
	}
	loaded, _ = fm.GetIndexFiles()
	if loaded != nil {
		t.Errorf("expected nil after ClearIndex, got %+v", loaded)
	}
}

// --- ListAll --- //
func TestListAll(t *testing.T) {
	dir := t.TempDir()
	fm := &file.FileManager{Root: dir}

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0o755)
	os.WriteFile(filepath.Join(dir, "subdir/b.txt"), []byte("y"), 0o644)

	all, err := fm.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 files, got %d", len(all))
	}
}

func TestFileManager_ErrorBranches(t *testing.T) {
	tmp := makeTempDir(t)
	fm := &file.FileManager{Root: tmp}

	// 1. CreateEntry/Write with nil BlockManager
	if _, err := fm.CreateEntry("a.txt"); err == nil {
		t.Error("expected error for CreateEntry with nil Blocks")
	}
	if err := fm.Write(file.Entry{Path: "x"}); err == nil {
		t.Error("expected error for Write with nil Blocks")
	}

	// 2. Exists on missing file
	if fm.Exists(filepath.Join(tmp, "nofile.txt")) {
		t.Error("expected Exists to return false")
	}

	// 3. StageFiles marshal error (use circular data to break json.Marshal)
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

	if err := fm.StageFiles([]file.Entry{{Path: "a"}}); err == nil {
		t.Error("expected StageFiles write failure")
	}

	// 4. StageFiles mkdir error — simulate by patching fsio.MkdirAll
	oldMkdir := fsio.MkdirAll
	fsio.MkdirAll = func(string, os.FileMode) error {
		return errors.New("fake mkdir error")
	}
	defer func() { fsio.MkdirAll = oldMkdir }()
	if err := fm.StageFiles([]file.Entry{}); err == nil {
		t.Error("expected mkdir failure")
	}

	// 5. ClearIndex with missing file
	if err := fm.ClearIndex(); err != nil {
		t.Error("ClearIndex should not error on missing file")
	}

	// 6. GetIndexFiles missing index.json
	if entries, err := fm.GetIndexFiles(); err != nil || entries != nil {
		t.Error("expected nil,nil for missing index.json")
	}

	// 7. GetIndexFiles bad JSON
	idx := filepath.Join(tmp, "index.json")
	os.WriteFile(idx, []byte("{ bad json"), 0o644)
	if _, err := fm.GetIndexFiles(); err == nil {
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
	fm := &file.FileManager{
		Root:   tmp,
		Blocks: makeBlockManager(t, tmp),
	}

	// CreateAllEntries with no files — should not panic
	_, _ = fm.CreateAllEntries()

	// CreateChangedEntries with empty index
	if _, err := fm.CreateChangedEntries(); err != nil {
		t.Errorf("unexpected error on empty index: %v", err)
	}
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
