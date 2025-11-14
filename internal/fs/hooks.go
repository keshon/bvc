package fs

import "os"

// Hooks used for testing (overridable)
var (
	open       = os.Open
	readFile   = os.ReadFile
	writeFile  = os.WriteFile
	stat       = os.Stat
	readDir    = os.ReadDir
	remove     = os.Remove
	rename     = os.Rename
	createTemp = os.CreateTemp
	mkdirAll   = os.MkdirAll
	isNotExist = os.IsNotExist
)

var exists = func(path string) bool {
	_, err := stat(path)
	return err == nil
}

var IsDir = func(path string) bool {
	fi, err := stat(path)
	return err == nil && fi.IsDir()
}

// getters and setters for test override
func GetOpen() func(string) (*os.File, error)    { return open }
func SetOpen(f func(string) (*os.File, error))   { open = f }
func GetReadFile() func(string) ([]byte, error)  { return readFile }
func SetReadFile(f func(string) ([]byte, error)) { readFile = f }
func GetWriteFile() func(string, []byte, os.FileMode) error {
	return writeFile
}
func SetWriteFile(f func(string, []byte, os.FileMode) error) {
	writeFile = f
}
func GetStat() func(string) (os.FileInfo, error)       { return stat }
func SetStat(f func(string) (os.FileInfo, error))      { stat = f }
func GetReadDir() func(string) ([]os.DirEntry, error)  { return readDir }
func SetReadDir(f func(string) ([]os.DirEntry, error)) { readDir = f }
func GetRemove() func(string) error                    { return remove }
func SetRemove(f func(string) error)                   { remove = f }
func GetRename() func(string, string) error            { return rename }
func SetRename(f func(string, string) error)           { rename = f }
func GetCreateTemp() func(string, string) (*os.File, error) {
	return createTemp
}
func SetCreateTemp(f func(string, string) (*os.File, error)) {
	createTemp = f
}
func GetMkdirAll() func(string, os.FileMode) error  { return mkdirAll }
func SetMkdirAll(f func(string, os.FileMode) error) { mkdirAll = f }
func GetIsNotExist() func(error) bool               { return isNotExist }
func SetIsNotExist(f func(error) bool)              { isNotExist = f }
