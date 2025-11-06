package fsio

import (
	"os"
)

// Hooks for filesystem operations
var (
	ReadFile       = os.ReadFile
	WriteFile      = os.WriteFile
	StatFile       = os.Stat
	ReadDir        = os.ReadDir
	Remove         = os.Remove
	Rename         = os.Rename
	CreateTempFile = os.CreateTemp
)
