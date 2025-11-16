package fs_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/keshon/bvc/internal/fs"
)

// Swaps hooks helpers

func fsOpenSwap(fn func(path string) (*os.File, error)) func() {
	old := fs.GetOpen()
	fs.SetOpen(fn)
	return func() { fs.SetOpen(old) }
}

func fsReadFileSwap(fn func(path string) ([]byte, error)) func() {
	old := fs.GetReadFile()
	fs.SetReadFile(fn)
	return func() { fs.SetReadFile(old) }
}

func fsWriteFileSwap(fn func(path string, data []byte, perm os.FileMode) error) func() {
	old := fs.GetWriteFile()
	fs.SetWriteFile(fn)
	return func() { fs.SetWriteFile(old) }
}

func fsStatSwap(fn func(path string) (os.FileInfo, error)) func() {
	old := fs.GetStat()
	fs.SetStat(fn)
	return func() { fs.SetStat(old) }
}

func fsReadDirSwap(fn func(path string) ([]os.DirEntry, error)) func() {
	old := fs.GetReadDir()
	fs.SetReadDir(fn)
	return func() { fs.SetReadDir(old) }
}

func fsRemoveSwap(fn func(path string) error) func() {
	old := fs.GetRemove()
	fs.SetRemove(fn)
	return func() { fs.SetRemove(old) }
}

func fsRenameSwap(fn func(old, new string) error) func() {
	old := fs.GetRename()
	fs.SetRename(fn)
	return func() { fs.SetRename(old) }
}

func fsCreateTempSwap(fn func(dir, pattern string) (io.WriteCloser, string, error)) func() {
	old := fs.GetCreateTemp()
	fs.SetCreateTemp(fn)
	return func() { fs.SetCreateTemp(old) }
}

func fsMkdirAllSwap(fn func(path string, perm os.FileMode) error) func() {
	old := fs.GetMkdirAll()
	fs.SetMkdirAll(fn)
	return func() { fs.SetMkdirAll(old) }
}

func fsIsNotExistSwap(fn func(err error) bool) func() {
	old := fs.GetIsNotExist()
	fs.SetIsNotExist(fn)
	return func() { fs.SetIsNotExist(old) }
}

// Tests
func TestOSFS_Open(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	// override hook
	fsOpenOrig := fsOpenSwap(func(path string) (*os.File, error) {
		called = true
		if path != "abc.txt" {
			t.Fatalf("expected path abc.txt, got %s", path)
		}
		return nil, errors.New("open-error")
	})
	defer fsOpenOrig()

	_, err := fsOverride.Open("abc.txt")
	if !called {
		t.Fatal("hook not called")
	}
	if err == nil || err.Error() != "open-error" {
		t.Fatalf("expected open-error, got %v", err)
	}
}

func TestOSFS_Stat(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	restore := fsStatSwap(func(path string) (os.FileInfo, error) {
		called = true
		return nil, errors.New("stat-failed")
	})
	defer restore()

	_, err := fsOverride.Stat("zzz")
	if !called {
		t.Fatal("expected stat hook to be called")
	}
	if err == nil || err.Error() != "stat-failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOSFS_ReadFile(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	restore := fsReadFileSwap(func(path string) ([]byte, error) {
		called = true
		return []byte("hello"), nil
	})
	defer restore()

	out, err := fsOverride.ReadFile("x")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("readFile hook not called")
	}
	if string(out) != "hello" {
		t.Fatalf("expected hello, got %s", out)
	}
}

func TestOSFS_WriteFile(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	restore := fsWriteFileSwap(func(path string, data []byte, perm os.FileMode) error {
		called = true
		if path != "aaa" || string(data) != "bbb" || perm != 0o644 {
			t.Fatalf("unexpected write args")
		}
		return nil
	})
	defer restore()

	err := fsOverride.WriteFile("aaa", []byte("bbb"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("writeFile hook not called")
	}
}

func TestOSFS_MkdirAll(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	restore := fsMkdirAllSwap(func(path string, perm os.FileMode) error {
		called = true
		if perm != 0o755 {
			t.Fatalf("unexpected perm")
		}
		return nil
	})
	defer restore()

	err := fsOverride.MkdirAll("dir123", 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("mkdirAll hook not called")
	}
}

func TestOSFS_Remove(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	restore := fsRemoveSwap(func(path string) error {
		called = true
		return nil
	})
	defer restore()

	err := fsOverride.Remove("qqq")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("remove hook not called")
	}
}

func TestOSFS_Rename(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	restore := fsRenameSwap(func(old, new string) error {
		called = true
		if old != "a" || new != "b" {
			t.Fatalf("unexpected rename args")
		}
		return nil
	})
	defer restore()

	err := fsOverride.Rename("a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("rename hook not called")
	}
}

func TestOSFS_CreateTempFile(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}

	restore := fsCreateTempSwap(func(dir, pattern string) (io.WriteCloser, string, error) {
		called = true
		if dir != "tmp" || pattern != "x*" {
			t.Fatalf("unexpected CreateTemp args: got dir=%q, pattern=%q", dir, pattern)
		}
		return nil, "", errors.New("tmp-failed")
	})
	defer restore()

	_, _, err := fsOverride.CreateTempFile("tmp", "x*")
	if err == nil || err.Error() != "tmp-failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("CreateTemp hook not called")
	}
}

func TestOSFS_IsNotExist(t *testing.T) {
	called := false
	fsOverride := &fs.OSFS{}
	errFake := errors.New("nope")

	restore := fsIsNotExistSwap(func(err error) bool {
		called = true
		return err == errFake
	})
	defer restore()

	if !fsOverride.IsNotExist(errFake) {
		t.Fatal("expected true")
	}
	if !called {
		t.Fatal("isNotExist not called")
	}
}

func TestOSFS_IsDir(t *testing.T) {
	tmp := t.TempDir()
	fsOverride := &fs.OSFS{}

	if !fsOverride.IsDir(tmp) {
		t.Fatalf("expected %s to be a dir", tmp)
	}
}

func TestOSFS_Exists(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "x")
	os.WriteFile(tmpFile, []byte("1"), 0o644)

	fsOverride := &fs.OSFS{}
	if !fsOverride.Exists(tmpFile) {
		t.Fatalf("expected file to exist")
	}
}
