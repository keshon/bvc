package fsio

import (
	"os"
)

// Hooks for filesystem operations
var (
	Open           = os.Open
	ReadFile       = os.ReadFile
	WriteFile      = os.WriteFile
	StatFile       = os.Stat
	ReadDir        = os.ReadDir
	Remove         = os.Remove
	Rename         = os.Rename
	CreateTempFile = os.CreateTemp
	MkdirAll       = os.MkdirAll
	IsNotExist     = os.IsNotExist
)
