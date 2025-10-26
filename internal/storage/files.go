package storage

import (
	"app/internal/config"
	"os"
	"path/filepath"
)

// listFiles returns all relevant file paths except the repository dir and binary itself.
func listFiles() ([]string, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	var paths []string
	err = filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path == config.RepoDir {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		abs, _ := filepath.Abs(path)
		if abs == exe {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths, err
}
