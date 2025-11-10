package file

import (
	"app/internal/config"
	"os"
	"path/filepath"
	"sort"
)

// ScanFilesInWorkingTree returns all user files in the working directory (excluding .bvc).
func (fc *FileContext) ScanFilesInWorkingTree() ([]string, error) {
	exe, _ := os.Executable()
	matcher := NewIgnore()

	var paths []string
	err := filepath.WalkDir(fc.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		clean := filepath.Clean(path)

		// Skip ignored directories
		if d.IsDir() {
			if d.Name() == config.RepoDir || matcher.Match(clean) {
				return filepath.SkipDir
			}
			return nil
		}

		if matcher.Match(clean) || clean == exe {
			return nil
		}

		paths = append(paths, clean)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}
