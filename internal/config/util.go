package config

import (
	"app/internal/fs"
	"os"
	"path/filepath"
)

// ResolveRepoRoot determines the actual repository root.
// It respects the .bvc-pointer file, if it exists.
func ResolveRepoRoot() string {
	root := RepoDir
	fs := fs.NewOSFS()

	if fi, err := fs.Stat(RepoPointerFile); err == nil && !fi.IsDir() {
		if data, err := fs.ReadFile(RepoPointerFile); err == nil {
			target := filepath.Clean(string(data))
			if filepath.IsAbs(target) {
				root = target
			} else {
				root = filepath.Join(".", target)
			}
		}
	}

	return root
}

// ResolveWorkingTreeRoot determines the working tree root by walking up.
// It traverses up the directory tree until it finds a .bvc directory or a .bvc-pointer file.
func ResolveWorkingTreeRoot() string {
	cwd, _ := os.Getwd()
	fs := fs.NewOSFS()

	for {
		bvcDir := filepath.Join(cwd, RepoDir)
		ptrFile := filepath.Join(cwd, RepoPointerFile)

		if fs.IsDir(bvcDir) || fs.Exists(ptrFile) {
			return cwd
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			break // reached filesystem root
		}
		cwd = parent
	}
	return "" // not found
}
