package file

import (
	"os"
	"path/filepath"
	"sort"
)

// ScanAllRepository returns slices of tracked, staged, and ignored files
// using the FS abstraction. Fully compatible with MemoryFS or OS FS.
// - tracked: files not ignored and not internal
// - staged: files already present in index.json
// - ignored: files matched by .bvc-ignore or defaults
func (fc *FileContext) ScanAllRepository() (tracked []string, staged []string, ignored []string, err error) {
	// physical paths (only for comparison, not for FS ops)
	repoAbs, _ := filepath.Abs(fc.RepoDir)
	exeAbs, _ := filepath.Abs(os.Args[0])

	matcher := NewIgnore(fc.WorkingTreeDir, fc.FS)

	// load index
	indexEntries, _ := fc.LoadIndex()
	indexSet := make(map[string]struct{})
	for _, e := range indexEntries {
		clean := filepath.ToSlash(filepath.Clean(e.Path))
		indexSet[clean] = struct{}{}
	}

	// --- Walk relative to working tree (MemoryFS safe) ---
	var walk func(rel string) error
	walk = func(rel string) error {
		entries, err := fc.FS.ReadDir(filepath.Join(fc.WorkingTreeDir, rel))
		if err != nil {
			return err
		}

		for _, e := range entries {
			childRel := filepath.Join(rel, e.Name())
			childAbs := filepath.Join(fc.WorkingTreeDir, childRel)

			info, _ := e.Info()
			if info == nil {
				continue
			}

			// skip repo dir (compare by abs)
			if childAbs == repoAbs {
				continue
			}

			// skip running binary (compare by abs)
			if childAbs == exeAbs {
				continue
			}

			// normalize rel for matcher
			relSlash := filepath.ToSlash(childRel)

			// directory
			if info.IsDir() {
				if matcher.Match(relSlash) {
					ignored = append(ignored, childAbs)
					continue
				}
				if err := walk(childRel); err != nil {
					return err
				}
				continue
			}

			// file
			if matcher.Match(relSlash) {
				ignored = append(ignored, childAbs)
				continue
			}

			if _, ok := indexSet[relSlash]; ok {
				staged = append(staged, childAbs)
			} else {
				tracked = append(tracked, childAbs)
			}
		}
		return nil
	}

	// walk starting from ""
	if err := walk(""); err != nil {
		return nil, nil, nil, err
	}

	sort.Strings(tracked)
	sort.Strings(staged)
	sort.Strings(ignored)

	return tracked, staged, ignored, nil
}
