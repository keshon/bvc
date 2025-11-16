package file_test

import (
	"app/internal/repo/store/block"
	"app/internal/repo/store/file"
	"path/filepath"
	"testing"
)

func TestRestoreFiles(t *testing.T) {
	fc, tmpDir := newTestFC(t)

	target := filepath.Join(tmpDir, "foo.txt")

	// Prepare block "hash1" with data "data"
	blockData := []byte("data")
	if err := fc.FS.WriteFile(
		filepath.Join(fc.BlockCtx.GetBlocksDir(), "hash1.bin"),
		blockData,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	// Prepare entry that references that block
	entry := file.Entry{
		Path:   target,
		Blocks: []block.BlockRef{{Hash: "hash1", Size: int64(len(blockData)), Offset: 0}},
	}

	// Call restore
	if err := fc.RestoreFilesToWorkingTree([]file.Entry{entry}, "test"); err != nil {
		t.Fatal(err)
	}

	// Verify result
	data, _ := fc.FS.ReadFile(target)
	if string(data) != "data" {
		t.Errorf("unexpected file content %q", data)
	}
}
