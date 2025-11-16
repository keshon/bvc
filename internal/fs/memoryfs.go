package fs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// MemoryFS is a pure in-memory filesystem for tests or lightweight storage.
type MemoryFS struct {
	files map[string][]byte
	dirs  map[string]struct{}
}

func NewMemoryFS() *MemoryFS {
	f := &MemoryFS{
		files: make(map[string][]byte),
		dirs:  make(map[string]struct{}),
	}
	f.dirs["/"] = struct{}{}
	f.dirs["."] = struct{}{}
	return f
}

// normalize paths
func clean(p string) string {
	if p == "" {
		return "."
	}
	return filepath.ToSlash(filepath.Clean(p))
}

func (f *MemoryFS) ensureDirExists(p string) error {
	p = clean(p)
	if _, ok := f.dirs[p]; !ok {
		return fs.ErrNotExist
	}
	return nil
}

// FS Interface Implementation

func (f *MemoryFS) Open(p string) (io.ReadSeekCloser, error) {
	p = clean(p)
	data, ok := f.files[p]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return &memReadSeekCloser{Reader: bytes.NewReader(data)}, nil
}

type memReadSeekCloser struct {
	*bytes.Reader
}

func (m *memReadSeekCloser) Close() error { return nil }

func (f *MemoryFS) ReadFile(p string) ([]byte, error) {
	p = clean(p)
	data, ok := f.files[p]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return append([]byte(nil), data...), nil
}

func (f *MemoryFS) WriteFile(p string, data []byte, perm os.FileMode) error {
	p = clean(p)
	dir := path.Dir(p)
	if err := f.ensureDirExists(dir); err != nil {
		return fmt.Errorf("write: dir %q does not exist", dir)
	}
	f.files[p] = append([]byte(nil), data...)
	return nil
}

func (f *MemoryFS) MkdirAll(p string, perm os.FileMode) error {
	p = clean(p)
	parts := strings.Split(p, "/")
	cur := ""
	for _, seg := range parts {
		if seg == "" || seg == "." {
			continue
		}
		cur = path.Join(cur, seg)
		if _, ok := f.dirs[cur]; !ok {
			f.dirs[cur] = struct{}{}
		}
	}
	return nil
}

func (f *MemoryFS) Remove(p string) error {
	p = clean(p)
	if _, ok := f.files[p]; ok {
		delete(f.files, p)
		return nil
	}
	if _, ok := f.dirs[p]; ok {
		delete(f.dirs, p)
		return nil
	}
	return fs.ErrNotExist
}

func (f *MemoryFS) Rename(oldp, newp string) error {
	oldp, newp = clean(oldp), clean(newp)

	// file rename
	if data, ok := f.files[oldp]; ok {
		dir := path.Dir(newp)
		if f.ensureDirExists(dir) != nil {
			return fs.ErrNotExist
		}
		delete(f.files, oldp)
		f.files[newp] = data
		return nil
	}

	// dir rename
	if _, ok := f.dirs[oldp]; ok {
		delete(f.dirs, oldp)
		f.dirs[newp] = struct{}{}
		return nil
	}

	return fs.ErrNotExist
}

func (f *MemoryFS) Stat(p string) (os.FileInfo, error) {
	p = clean(p)
	if data, ok := f.files[p]; ok {
		return &fakeInfo{name: filepath.Base(p), size: int64(len(data)), dir: false}, nil
	}
	if _, ok := f.dirs[p]; ok {
		return &fakeInfo{name: filepath.Base(p), dir: true}, nil
	}
	return nil, fs.ErrNotExist
}

func (f *MemoryFS) ReadDir(p string) ([]os.DirEntry, error) {
	p = clean(p)
	if _, ok := f.dirs[p]; !ok {
		return nil, fs.ErrNotExist
	}

	var out []os.DirEntry
	prefix := p
	if prefix != "/" && prefix != "." {
		prefix += "/"
	}

	seen := map[string]bool{}

	// dirs first
	for dp := range f.dirs {
		if strings.HasPrefix(dp, prefix) {
			rest := strings.TrimPrefix(dp, prefix)
			name := strings.Split(rest, "/")[0]
			if name != "" && name != "." && !seen[name] {
				seen[name] = true
				out = append(out, fakeDirEntry{name: name, isDir: true})
			}
		}
	}

	// then files
	for fp := range f.files {
		if strings.HasPrefix(fp, prefix) {
			rest := strings.TrimPrefix(fp, prefix)
			name := strings.Split(rest, "/")[0]
			if name != "" && !seen[name] {
				seen[name] = true
				out = append(out, fakeDirEntry{name: name, isDir: false})
			}
		}
	}

	return out, nil
}

func (f *MemoryFS) CreateTempFile(dir, pattern string) (io.WriteCloser, string, error) {
	if err := f.ensureDirExists(clean(dir)); err != nil {
		return nil, "", err
	}

	tmpName := filepath.Join(dir, pattern+"-tmp")
	buf := &bytes.Buffer{}

	wc := &memWriteCloser{
		buf: buf,
		onClose: func() {
			f.files[clean(tmpName)] = buf.Bytes()
		},
	}
	return wc, tmpName, nil
}

type memWriteCloser struct {
	buf     *bytes.Buffer
	onClose func()
}

func (m *memWriteCloser) Write(p []byte) (int, error) { return m.buf.Write(p) }
func (m *memWriteCloser) Close() error {
	if m.onClose != nil {
		m.onClose()
	}
	return nil
}

func (f *MemoryFS) IsNotExist(err error) bool { return errors.Is(err, fs.ErrNotExist) }
func (f *MemoryFS) IsDir(p string) bool       { _, ok := f.dirs[clean(p)]; return ok }
func (f *MemoryFS) Exists(p string) bool {
	p = clean(p)
	_, f1 := f.files[p]
	_, d1 := f.dirs[p]
	return f1 || d1
}

// Helpers

type fakeInfo struct {
	name string
	size int64
	dir  bool
}

func (f *fakeInfo) Name() string       { return f.name }
func (f *fakeInfo) Size() int64        { return f.size }
func (f *fakeInfo) Mode() fs.FileMode  { return 0o644 }
func (f *fakeInfo) ModTime() time.Time { return time.Time{} }
func (f *fakeInfo) IsDir() bool        { return f.dir }
func (f *fakeInfo) Sys() interface{}   { return nil }

type fakeDirEntry struct {
	name  string
	isDir bool
}

func (d fakeDirEntry) Name() string               { return d.name }
func (d fakeDirEntry) IsDir() bool                { return d.isDir }
func (d fakeDirEntry) Type() fs.FileMode          { return 0 }
func (d fakeDirEntry) Info() (os.FileInfo, error) { return &fakeInfo{name: d.name, dir: d.isDir}, nil }
