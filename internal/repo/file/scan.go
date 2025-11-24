package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/keshon/bvc/internal/repo/ignore"
)

// ScanAllRepository returns slices of tracked, staged, and ignored files
// using the FS abstraction. Fully compatible with MemoryFS or OS FS.
// - tracked: files not ignored and not internal
// - staged: files already present in index.json
// - ignored: files matched by .bvc-ignore or defaults
func (fc *FileContext) ScanAllRepository() (tracked, staged, ignored []string, err error) {
	// абсолютный путь к .bvc внутри рабочей директории
	repoAbs := filepath.Join(fc.WorkingTreeDir, fc.RepoDir)

	// имя текущего бинарника без пути и расширения
	binBase := filepath.Base(os.Args[0])
	binName := strings.TrimSuffix(binBase, filepath.Ext(binBase))

	matcher := ignore.NewIgnore(fc.WorkingTreeDir, fc.FS)

	// load staged entries
	indexEntries, _ := fc.LoadIndex()
	indexSet := make(map[string]struct{}, len(indexEntries))
	for _, e := range indexEntries {
		indexSet[filepath.ToSlash(filepath.Clean(e.Path))] = struct{}{}
	}

	// рекурсивный обход
	var walk func(rel string) error
	walk = func(rel string) error {
		dirAbs := filepath.Join(fc.WorkingTreeDir, rel)
		entries, err := fc.FS.ReadDir(dirAbs)
		if err != nil {
			return err
		}

		for _, e := range entries {
			childRel := filepath.Join(rel, e.Name())
			childAbs := filepath.Join(fc.WorkingTreeDir, childRel)

			info, err := e.Info()
			if err != nil {
				fmt.Println(err)
			}
			if info == nil {
				continue
			}

			// skip repo dir
			if childAbs == repoAbs {
				continue
			}

			// skip current binary
			if !info.IsDir() {
				fileName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
				if fileName == binName {
					continue
				}
			}

			relSlash := filepath.ToSlash(childRel)

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

			// файл
			if matcher.Match(relSlash) {
				ignored = append(ignored, childAbs)
			} else if _, ok := indexSet[relSlash]; ok {
				staged = append(staged, childAbs)
			} else {
				tracked = append(tracked, childAbs)
			}
		}
		return nil
	}

	if err := walk(""); err != nil {
		return nil, nil, nil, err
	}

	sort.Strings(tracked)
	sort.Strings(staged)
	sort.Strings(ignored)
	return tracked, staged, ignored, nil
}
