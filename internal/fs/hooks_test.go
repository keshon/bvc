package fs_test

import (
	"app/internal/fs"
	"errors"
	"io"
	"os"
	"testing"
)

func TestHookOverrides(t *testing.T) {
	// open hook
	orig := fs.GetOpen()
	defer fs.SetOpen(orig)

	called := false
	fs.SetOpen(func(path string) (*os.File, error) {
		called = true
		return nil, errors.New("open-error")
	})

	_, err := fs.GetOpen()("x")
	if !called {
		t.Fatal("Open hook not called")
	}
	if err == nil || err.Error() != "open-error" {
		t.Fatalf("unexpected error: %v", err)
	}

	// readFile hook
	origRF := fs.GetReadFile()
	defer fs.SetReadFile(origRF)

	called = false
	fs.SetReadFile(func(path string) ([]byte, error) {
		called = true
		return []byte("ok"), nil
	})
	out, err := fs.GetReadFile()("y")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "ok" {
		t.Fatalf("expected ok, got %s", out)
	}
	if !called {
		t.Fatal("ReadFile hook not called")
	}

	// writeFile hook
	origWF := fs.GetWriteFile()
	defer fs.SetWriteFile(origWF)

	called = false
	fs.SetWriteFile(func(path string, data []byte, perm os.FileMode) error {
		called = true
		if path != "a" || string(data) != "b" || perm != 0o644 {
			t.Fatalf("unexpected args")
		}
		return nil
	})
	err = fs.GetWriteFile()("a", []byte("b"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("WriteFile hook not called")
	}

	// stat hook
	origStat := fs.GetStat()
	defer fs.SetStat(origStat)

	called = false
	fs.SetStat(func(path string) (os.FileInfo, error) {
		called = true
		return nil, errors.New("stat-error")
	})
	_, err = fs.GetStat()("z")
	if !called {
		t.Fatal("Stat hook not called")
	}
	if err == nil || err.Error() != "stat-error" {
		t.Fatalf("unexpected error: %v", err)
	}

	// mkdirAll hook
	origMk := fs.GetMkdirAll()
	defer fs.SetMkdirAll(origMk)

	called = false
	fs.SetMkdirAll(func(path string, perm os.FileMode) error {
		called = true
		if perm != 0o755 {
			t.Fatalf("unexpected perm")
		}
		return nil
	})
	err = fs.GetMkdirAll()("dir", 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("MkdirAll hook not called")
	}

	// remove hook
	origRm := fs.GetRemove()
	defer fs.SetRemove(origRm)

	called = false
	fs.SetRemove(func(path string) error {
		called = true
		return nil
	})
	err = fs.GetRemove()("file")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Remove hook not called")
	}

	// rename hook
	origRen := fs.GetRename()
	defer fs.SetRename(origRen)

	called = false
	fs.SetRename(func(old, new string) error {
		called = true
		if old != "x" || new != "y" {
			t.Fatalf("unexpected args")
		}
		return nil
	})
	err = fs.GetRename()("x", "y")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Rename hook not called")
	}

	// createTemp hook
	origTmp := fs.GetCreateTemp()
	defer fs.SetCreateTemp(origTmp)

	called = false
	fs.SetCreateTemp(func(dir, pattern string) (io.WriteCloser, string, error) {
		called = true
		return nil, "tmpfile", errors.New("tmp-err")
	})
	_, _, err = fs.GetCreateTemp()("d", "p")
	if err == nil || err.Error() != "tmp-err" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("CreateTemp hook not called")
	}

	// isNotExist hook
	origNE := fs.GetIsNotExist()
	defer fs.SetIsNotExist(origNE)

	called = false
	fs.SetIsNotExist(func(err error) bool {
		called = true
		return true
	})
	if !fs.GetIsNotExist()(errors.New("x")) {
		t.Fatal("expected true from IsNotExist hook")
	}
	if !called {
		t.Fatal("IsNotExist hook not called")
	}
}
