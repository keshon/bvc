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
	entry := file.Entry{
		Path:   target,
		Blocks: []block.BlockRef{{Hash: "hash1", Size: 3}},
	}

	// TODO: finish the test

	if err := fc.RestoreFilesToWorkingTree([]file.Entry{entry}, "test"); err != nil {
		t.Fatal(err)
	}

	data, _ := fc.FS.ReadFile(target)
	if string(data) != "data" {
		t.Errorf("unexpected file content %q", data)
	}
}
