package file_test

import (
	"app/internal/fs"
	"app/internal/repo/store/file"
	"os"
	"path/filepath"
	"testing"
)

func TestScanFiles(t *testing.T) {
	tmp := t.TempDir()

	// write real files on disk
	_ = os.WriteFile(filepath.Join(tmp, "foo.txt"), []byte("ok"), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, "ignore.me"), []byte("ok"), 0o644)

	// create .bvc-ignore
	_ = os.WriteFile(filepath.Join(tmp, ".bvc-ignore"), []byte("ignore.me\n"), 0o644)

	// configure fc to point at tmp dir and to use a mock FS for index loads
	fs := fs.NewMemoryFS()
	fc := &file.FileContext{FS: fs, WorkingTreeDir: tmp, RepoDir: tmp}

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
	tmp := t.TempDir()
	fs := fs.NewMemoryFS()
	fc := &file.FileContext{FS: fs, WorkingTreeDir: tmp, RepoDir: tmp}

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
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "sub/dir"), 0o755)
	os.WriteFile(filepath.Join(tmp, "sub/dir/a.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(tmp, "sub/b.txt"), []byte("y"), 0o644)

	fs := fs.NewMemoryFS()
	fc := &file.FileContext{FS: fs, WorkingTreeDir: tmp, RepoDir: tmp}

	tracked, _, _, err := fc.ScanAllRepository()
	if err != nil {
		t.Fatal(err)
	}
	if len(tracked) != 2 {
		t.Errorf("expected 2 tracked files, got %d", len(tracked))
	}
}
