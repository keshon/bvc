package fs_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/keshon/bvc/internal/fs"
)

func TestMemoryFS_WriteReadFile(t *testing.T) {
	m := fs.NewMemoryFS()

	// Create dirs first
	if err := m.MkdirAll("dir/sub", 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("hello world")
	if err := m.WriteFile("dir/sub/file.txt", content, 0o644); err != nil {
		t.Fatal(err)
	}

	read, err := m.ReadFile("dir/sub/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(read, content) {
		t.Fatalf("expected %q, got %q", content, read)
	}
}

func TestMemoryFS_WriteFileNonExistentDir(t *testing.T) {
	m := fs.NewMemoryFS()
	err := m.WriteFile("nope/file.txt", []byte("x"), 0o644)
	if err == nil {
		t.Fatal("expected error writing to non-existent dir")
	}
}

func TestMemoryFS_OpenAndClose(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("d", 0o755)
	m.WriteFile("d/f", []byte("abc"), 0o644)

	f, err := m.Open("d/f")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	buf := make([]byte, 3)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if n != 3 || string(buf) != "abc" {
		t.Fatalf("unexpected read %q", buf)
	}
}

func TestMemoryFS_Remove(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("d", 0o755)
	m.WriteFile("d/f", []byte("x"), 0o644)

	if !m.Exists("d/f") {
		t.Fatal("file should exist")
	}

	if err := m.Remove("d/f"); err != nil {
		t.Fatal(err)
	}
	if m.Exists("d/f") {
		t.Fatal("file should be removed")
	}

	// remove non-existent
	if err := m.Remove("missing"); !errors.Is(err, os.ErrNotExist) && !m.IsNotExist(err) {
		t.Fatal("expected not-exist error")
	}
}

func TestMemoryFS_RenameFileAndDir(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("dir/sub", 0o755)
	m.WriteFile("dir/f", []byte("data"), 0o644)

	// File rename
	if err := m.Rename("dir/f", "dir/f2"); err != nil {
		t.Fatal(err)
	}
	if m.Exists("dir/f") || !m.Exists("dir/f2") {
		t.Fatal("file rename failed")
	}

	// Dir rename
	if err := m.Rename("dir/sub", "dir/sub2"); err != nil {
		t.Fatal(err)
	}
	if m.Exists("dir/sub") || !m.Exists("dir/sub2") {
		t.Fatal("dir rename failed")
	}

	// Rename non-existent
	if err := m.Rename("nope", "new"); !m.IsNotExist(err) {
		t.Fatal("expected not-exist error")
	}
}

func TestMemoryFS_StatAndIsDir(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("a/b", 0o755)
	m.WriteFile("a/b/f.txt", []byte("x"), 0o644)

	info, err := m.Stat("a/b/f.txt")
	if err != nil || info.IsDir() {
		t.Fatal("expected file info")
	}

	info2, err := m.Stat("a/b")
	if err != nil || !info2.IsDir() {
		t.Fatal("expected dir info")
	}

	if !m.IsDir("a/b") || m.IsDir("a/b/f.txt") && !info.IsDir() {
		t.Fatal("IsDir mismatch")
	}

	if _, err := m.Stat("missing"); !m.IsNotExist(err) {
		t.Fatal("expected not-exist error")
	}
}

func TestMemoryFS_ReadDir(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("root/a", 0o755)
	m.MkdirAll("root/b", 0o755)
	m.WriteFile("root/f1.txt", []byte("x"), 0o644)
	m.WriteFile("root/a/f2.txt", []byte("y"), 0o644)

	entries, err := m.ReadDir("root")
	if err != nil {
		t.Fatal(err)
	}

	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = e.IsDir()
	}

	expected := map[string]bool{"a": true, "b": true, "f1.txt": false}
	for k, v := range expected {

		isDir, ok := names[k]
		if !ok || isDir != v {
			t.Fatalf("expected %s=%v, got %v", k, v, isDir)
		}
	}

	if _, err := m.ReadDir("missing"); !m.IsNotExist(err) {
		t.Fatal("expected not-exist error")
	}
}

func TestMemoryFS_CreateTempFile(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("tmp", 0o755)

	wc, name, err := m.CreateTempFile("tmp", "x")
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("abc")
	if _, err := wc.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := wc.Close(); err != nil {
		t.Fatal(err)
	}

	read, err := m.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(read, data) {
		t.Fatalf("expected %q, got %q", data, read)
	}
}

func TestMemoryFS_Exists(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("d", 0o755)
	m.WriteFile("d/f", []byte("x"), 0o644)

	if !m.Exists("d") || !m.Exists("d/f") {
		t.Fatal("expected d and d/f to exist")
	}
	if m.Exists("missing") {
		t.Fatal("unexpected exists true")
	}
}

func TestMemoryFS_PathNormalization(t *testing.T) {
	m := fs.NewMemoryFS()
	m.MkdirAll("a/b", 0o755)
	m.WriteFile("a/b/f", []byte("x"), 0o644)

	// Use weird paths
	if !m.Exists("a/./b/../b/f") {
		t.Fatal("path normalization failed")
	}
	if !m.IsDir("a/./b/../b") {
		t.Fatal("path normalization failed for dir")
	}
}
