package file

import (
	"path/filepath"
	"sort"
)

// ScanAllRepository returns slices of tracked, staged, and ignored files
// using the FS abstraction. Fully compatible with MemoryFS or OS FS.
// - tracked: files not ignored and not internal
// - staged: files already present in index.json
// - ignored: files matched by .bvc-ignore or defaults
func (fc *FileContext) ScanAllRepository() (tracked []string, staged []string, ignored []string, err error) {
	matcher := NewIgnore(fc.WorkingTreeDir, fc.FS)

	// Load staged entries from index
	indexEntries, _ := fc.LoadIndex()
	indexSet := make(map[string]struct{}, len(indexEntries))
	for _, e := range indexEntries {
		indexSet[filepath.ToSlash(filepath.Clean(e.Path))] = struct{}{}
	}

	// Internal helper: recursive FS walk
	var walk func(path string) error
	walk = func(path string) error {
		entries, err := fc.FS.ReadDir(path)
		if err != nil {
			return err
		}

		for _, e := range entries {
			p := filepath.Join(path, e.Name())
			info, _ := e.Info()

			// Skip internal repo directory
			if info.IsDir() && filepath.Clean(p) == filepath.Clean(fc.RepoDir) {
				continue
			}

			// Skip ignored directories
			if info.IsDir() && matcher.Match(p) {
				ignored = append(ignored, p)
				continue
			}

			// Recurse into directories
			if info.IsDir() {
				if err := walk(p); err != nil {
					return err
				}
				continue
			}

			// File: normalize path relative to working tree
			relPath, err := filepath.Rel(fc.WorkingTreeDir, p)
			if err != nil {
				relPath = p
			}
			relPath = filepath.ToSlash(relPath)

			// Decide tracked/staged/ignored
			if matcher.Match(relPath) {
				ignored = append(ignored, p)
			} else if _, ok := indexSet[relPath]; ok {
				staged = append(staged, p)
			} else {
				tracked = append(tracked, p)
			}
		}

		return nil
	}

	if err := walk(fc.WorkingTreeDir); err != nil {
		return nil, nil, nil, err
	}

	// Sort for determinism
	sort.Strings(tracked)
	sort.Strings(staged)
	sort.Strings(ignored)

	return tracked, staged, ignored, nil
}
