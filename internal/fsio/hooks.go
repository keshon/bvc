package fsio

import (
	"os"
)

// Hooks for filesystem operations
// used for testing
var (
	Open           = os.Open
	ReadFile       = os.ReadFile
	WriteFile      = os.WriteFile
	StatFile       = os.Stat
	Stat           = os.Stat
	ReadDir        = os.ReadDir
	Remove         = os.Remove
	Rename         = os.Rename
	CreateTempFile = os.CreateTemp
	CreateTemp     = os.CreateTemp
	MkdirAll       = os.MkdirAll
	IsNotExist     = os.IsNotExist
	Exists         = func(path string) bool { _, err := StatFile(path); return err == nil }
	IsDir          = func(path string) bool { fi, err := StatFile(path); return err == nil && fi.IsDir() }
)
