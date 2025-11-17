package file_test

import (
	"path/filepath"
	"testing"

	"github.com/keshon/bvc/internal/repo/store/block"
	"github.com/keshon/bvc/internal/repo/store/file"
)

func TestRestoreFiles(t *testing.T) {
	fc, tmpDir := newTestFC(t)

	target := filepath.Join(tmpDir, "foo.txt")

	// Prepare block "hash1" with data "data"
	blockData := []byte("data")
	if err := fc.FS.WriteFile(
		filepath.Join(fc.BlockCtx.BlocksDir(), "hash1.bin"),
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
