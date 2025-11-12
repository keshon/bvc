package file

import (
	"app/internal/config"
	"os"
	"path/filepath"
	"sort"
)

// ScanFilesInWorkingTree returns two slices of file paths: tracked and ignored file paths (.bvc-ignore and .bvc/).
func (fc *FileContext) ScanFilesInWorkingTree() ([]string, []string, error) {
	exe, _ := os.Executable()
	matcher := NewIgnore()

	var ignored []string
	var tracked []string

	err := filepath.WalkDir(fc.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		clean := filepath.Clean(path)

		// skip repo internal dir (.bvc or similar)
		if d.IsDir() {
			if d.Name() == config.RepoDir {
				return filepath.SkipDir
			}
			if matcher.Match(clean) {
				ignored = append(ignored, clean)
				return filepath.SkipDir
			}
			return nil
		}

		if clean == exe {
			return nil
		}

		// normalize path to relative to root
		relPath, err := filepath.Rel(fc.Root, clean)
		if err != nil {
			relPath = clean
		}

		relPath = filepath.ToSlash(relPath)

		// split into tracked and ignored
		if matcher.Match(relPath) {
			ignored = append(ignored, clean)
		} else {
			tracked = append(tracked, clean)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	sort.Strings(tracked)
	sort.Strings(ignored)
	return tracked, ignored, nil
}
