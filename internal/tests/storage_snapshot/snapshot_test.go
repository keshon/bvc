package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
)

// helpers
func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bvc-snapshot-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir
}

func makeBlockManager(t *testing.T, dir string) *block.BlockManager {
	t.Helper()
	return &block.BlockManager{Root: dir}
}

func makeFileManager(t *testing.T, root string, bm *block.BlockManager) *file.FileManager {
	t.Helper()
	return &file.FileManager{Root: root, Blocks: bm}
}

func makeSnapshotManager(t *testing.T, root string, fm *file.FileManager, bm *block.BlockManager) *snapshot.SnapshotManager {
	t.Helper()
	return &snapshot.SnapshotManager{Root: root, Files: fm, Blocks: bm}
}

// --- Test SnapshotManager CreateCurrent + Create + Save/Load/List --- //
func TestSnapshotManagerWorkflow(t *testing.T) {
	root := makeTempDir(t)
	defer os.RemoveAll(root)

	// setup managers
	bm := makeBlockManager(t, filepath.Join(root, "blocks"))
	fm := makeFileManager(t, filepath.Join(root, "files"), bm)
	sm := makeSnapshotManager(t, filepath.Join(root, "snapshots"), fm, bm)

	// create test file
	filePath := filepath.Join(root, "test.txt")
	content := []byte("snapshot test content")
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// CreateCurrent fileset
	fs1, err := sm.CreateCurrent()
	if err != nil {
		t.Fatalf("CreateCurrent failed: %v", err)
	}
	if len(fs1.Files) != 1 {
		t.Fatalf("expected 1 file in fileset, got %d", len(fs1.Files))
	}

	// Create fileset explicitly from entries
	entry, err := fm.CreateEntry(filePath)
	if err != nil {
		t.Fatalf("CreateEntry failed: %v", err)
	}
	fs2, err := sm.Create([]file.Entry{entry})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if fs2.ID == "" {
		t.Fatal("fileset ID should not be empty")
	}

	// Save and Load
	if err := sm.Save(fs2); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := sm.Load(fs2.ID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.ID != fs2.ID {
		t.Errorf("loaded fileset ID mismatch: got %s, want %s", loaded.ID, fs2.ID)
	}

	// List filesets
	list, err := sm.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 fileset in List, got %d", len(list))
	}

	// WriteAndSave stores blocks and saves metadata
	fs3 := &snapshot.Fileset{
		ID:    "manual-id",
		Files: []file.Entry{entry},
	}
	if err := sm.WriteAndSave(fs3); err != nil {
		t.Fatalf("WriteAndSave failed: %v", err)
	}

	// verify blocks exist
	for _, b := range entry.Blocks {
		path := filepath.Join(bm.Root, b.Hash+".bin")
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected block file %s to exist: %v", path, err)
		} else if info.Size() != b.Size {
			t.Errorf("block size mismatch for %s: got %d, want %d", path, info.Size(), b.Size)
		}
	}
}

// --- Test HashFileset determinism --- //
func TestHashFilesetDeterminism(t *testing.T) {
	entry1 := file.Entry{
		Path: "a.txt",
		Blocks: []block.BlockRef{
			{Hash: "111", Size: 1, Offset: 0},
			{Hash: "222", Size: 2, Offset: 1},
		},
	}
	entry2 := file.Entry{
		Path: "b.txt",
		Blocks: []block.BlockRef{
			{Hash: "333", Size: 3, Offset: 0},
		},
	}
	fs1 := []file.Entry{entry1, entry2}
	fs2 := []file.Entry{entry2, entry1}

	hash1 := snapshot.HashFileset(fs1)
	hash2 := snapshot.HashFileset(fs2)

	if hash1 != hash2 {
		t.Errorf("HashFileset should be deterministic regardless of entry order: %s vs %s", hash1, hash2)
	}
}
