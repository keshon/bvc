package fs_test

import (
	"app/internal/fs"
	"os"
)

// Each of these swaps a hook with a test function
// and returns a restore() function to reset the hook.

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

func fsCreateTempSwap(fn func(dir, pattern string) (*os.File, error)) func() {
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
