package file

import (
	"app/internal/config"
	"os"
	"path/filepath"
	"sort"
)

// ScanFilesInWorkingTree returns three slices of file paths: tracked, staged, and ignored.
// - tracked: files not ignored and not internal
// - staged: files already present in index.json
// - ignored: files matched by .bvc-ignore or defaults
func (fc *FileContext) ScanFilesInWorkingTree() (tracked []string, staged []string, ignored []string, err error) {
	exe, _ := os.Executable()
	matcher := NewIgnore()

	// Load staged entries (index)
	indexEntries, _ := fc.LoadIndex()
	indexSet := make(map[string]struct{}, len(indexEntries))
	for _, e := range indexEntries {
		indexSet[filepath.ToSlash(filepath.Clean(e.Path))] = struct{}{}
	}

	err = filepath.WalkDir(fc.Root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		clean := filepath.Clean(path)

		// Skip internal repo directory (.bvc, etc.)
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

		// Skip current binary
		if clean == exe {
			return nil
		}

		// Normalize to relative to repo root
		relPath, err := filepath.Rel(fc.Root, clean)
		if err != nil {
			relPath = clean
		}
		relPath = filepath.ToSlash(relPath)

		// Check index membership
		_, inIndex := indexSet[relPath]

		// Decide where to put this path
		if matcher.Match(relPath) {
			ignored = append(ignored, clean)
		} else if inIndex {
			staged = append(staged, clean)
		} else {
			tracked = append(tracked, clean)
		}

		return nil
	})

	if err != nil {
		return nil, nil, nil, err
	}

	sort.Strings(tracked)
	sort.Strings(staged)
	sort.Strings(ignored)

	return tracked, staged, ignored, nil
}
