package file_test

import (
	"path/filepath"
	"testing"
)

func TestScanFiles(t *testing.T) {
	fc, tmpDir := newTestFC(t)

	// create working tree files using fc.FS
	fc.FS.WriteFile(filepath.Join(tmpDir, "foo.txt"), []byte("ok"), 0o644)
	fc.FS.WriteFile(filepath.Join(tmpDir, "ignore.me"), []byte("ok"), 0o644)

	// create .bvc-ignore
	fc.FS.WriteFile(filepath.Join(tmpDir, ".bvc-ignore"), []byte("ignore.me\n"), 0o644)

	tracked, _, ignored, err := fc.ScanAllRepository()
	if err != nil {
		t.Fatal(err)
	}

	if len(tracked) == 0 {
		t.Error("expected tracked files, got none")
	}
	if len(ignored) == 0 {
		t.Error("expected ignored files, got none")
	}
}

func TestScanEmptyRepo(t *testing.T) {
	fc, _ := newTestFC(t)

	// no files created at all; just scan empty repo

	tracked, staged, ignored, err := fc.ScanAllRepository()
	if err != nil {
		t.Fatal(err)
	}

	if len(tracked) != 0 {
		t.Errorf("expected 0 tracked files, got %d", len(tracked))
	}
	if len(staged) != 0 {
		t.Errorf("expected 0 staged files, got %d", len(staged))
	}
	if len(ignored) != 0 {
		t.Errorf("expected 0 ignored files, got %d", len(ignored))
	}
}

func TestScanNestedDirs(t *testing.T) {
	fc, tmpDir := newTestFC(t)

	// create nested directories and files in the MemoryFS via fc.FS
	fc.FS.MkdirAll(filepath.Join(tmpDir, "sub/dir"), 0o755)
	fc.FS.WriteFile(filepath.Join(tmpDir, "sub/dir/a.txt"), []byte("x"), 0o644)
	fc.FS.WriteFile(filepath.Join(tmpDir, "sub/b.txt"), []byte("y"), 0o644)

	tracked, _, _, err := fc.ScanAllRepository()
	if err != nil {
		t.Fatal(err)
	}

	if len(tracked) != 2 {
		t.Errorf("expected 2 tracked files, got %d", len(tracked))
	}
}
