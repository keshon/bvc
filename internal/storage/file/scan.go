package file

import (
	"app/internal/config"
	"os"
	"path/filepath"
	"sort"
)

// ListAll returns all user files in the working directory (excluding .bvc).
func (fm *FileManager) ListAll() ([]string, error) {
	exe, _ := os.Executable()
	var paths []string

	err := filepath.WalkDir(fm.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == config.RepoDir {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		abs, _ := filepath.Abs(path)
		if abs == exe {
			return nil
		}
		paths = append(paths, abs)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort for deterministic order
	sort.Strings(paths)
	return paths, nil
}
