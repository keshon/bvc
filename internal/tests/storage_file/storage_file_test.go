package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"app/internal/storage/block"
	"app/internal/storage/file"
)

func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bvc-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir
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

// // --- CreateAllEntries / CreateChangedEntries --- //
// func TestCreateAllAndChangedEntries(t *testing.T) {
// 	dir := makeTempDir(t)
// 	defer os.RemoveAll(dir)

// 	fm := &file.FileManager{Root: dir, Blocks: makeBlockManager(t, dir)}

// 	filePath := filepath.Join(dir, "file.txt")
// 	os.WriteFile(filePath, []byte("content"), 0o644)

// 	// Stage initial file
// 	entry, _ := fm.CreateEntry(filePath)
// 	fm.StageFiles([]file.Entry{entry})

// 	all, err := fm.CreateAllEntries()
// 	if err != nil {
// 		t.Fatalf("CreateAllEntries failed: %v", err)
// 	}
// 	if len(all) == 0 {
// 		t.Errorf("expected at least 1 entry in CreateAllEntries")
// 	}

// 	changed, err := fm.CreateChangedEntries()
// 	if err != nil {
// 		t.Fatalf("CreateChangedEntries failed: %v", err)
// 	}
// 	if len(changed) == 0 {
// 		t.Errorf("expected at least 1 entry in CreateChangedEntries")
// 	}
// }

// // --- Restore --- //
// func TestRestore(t *testing.T) {
// 	dir := makeTempDir(t)
// 	defer os.RemoveAll(dir)

// 	fm := &file.FileManager{Root: dir, Blocks: makeBlockManager(t, dir)}

// 	filePath := filepath.Join(dir, "file.txt")
// 	os.WriteFile(filePath, []byte("data"), 0o644)

// 	entry, _ := fm.CreateEntry(filePath)
// 	// first remove the file to force restore
// 	os.Remove(filePath)

// 	if err := fm.Restore([]file.Entry{entry}, "test"); err != nil {
// 		t.Fatalf("Restore failed: %v", err)
// 	}

// 	if _, err := os.Stat(filePath); err != nil {
// 		t.Errorf("restored file missing: %v", err)
// 	}
// }

// --- ListAll --- //
func TestListAll(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	fm := &file.FileManager{Root: dir}

	os.WriteFile("a.txt", []byte("x"), 0o644)
	os.Mkdir("subdir", 0o755)
	os.WriteFile("subdir/b.txt", []byte("y"), 0o644)

	all, err := fm.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 files, got %d", len(all))
	}
}
