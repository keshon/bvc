package file

import (
	"app/internal/fsio"
	"app/internal/progress"
	"app/internal/util"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Restore rebuilds files from entries (e.g., from a snapshot).
func (fc *FileContext) Restore(entries []Entry, label string) error {
	if fc.Blocks == nil {
		return fmt.Errorf("no BlockManager attached")
	}

	exe := filepath.Base(os.Args[0])
	bar := progress.NewProgress(len(entries), fmt.Sprintf("Restoring %s", label))
	defer bar.Finish()

	// Build valid file map for pruning
	valid := make(map[string]bool, len(entries))
	for _, e := range entries {
		valid[filepath.Clean(e.Path)] = true
	}

	// Restore files in parallel
	err := util.Parallel(entries, util.WorkerCount()*2, func(e Entry) error {
		if filepath.Base(e.Path) == exe {
			bar.Increment()
			return nil
		}
		if err := fc.restoreFile(e); err != nil {
			fmt.Printf("\nWarning: %v\n", err)
		}
		bar.Increment()
		return nil
	})

	// Remove files not in snapshot
	fc.pruneUntrackedFiles(valid, exe)
	return err
}

func (fc *FileContext) restoreFile(e Entry) error {
	if err := fsio.MkdirAll(filepath.Dir(e.Path), 0o755); err != nil {
		return err
	}

	tmp, err := fsio.CreateTempFile(filepath.Dir(e.Path), "tmp-*")
	if err != nil {
		return err
	}
	defer fsio.Remove(tmp.Name())
	defer tmp.Close()

	writer := bufio.NewWriterSize(tmp, 4*1024*1024)
	for _, b := range e.Blocks {
		data, err := fc.Blocks.Read(b.Hash)
		if err != nil {
			return fmt.Errorf("missing block %s for %s", b.Hash, e.Path)
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}
	writer.Flush()
	tmp.Sync()
	tmp.Close()

	return fsio.Rename(tmp.Name(), e.Path)
}

func (fc *FileContext) pruneUntrackedFiles(valid map[string]bool, exe string) {
	var dirs []string
	filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(path, fc.Root) {
				return filepath.SkipDir
			}
			dirs = append(dirs, path)
			return nil
		}
		if !valid[filepath.Clean(path)] && filepath.Base(path) != exe {
			_ = fsio.Remove(path)
		}
		return nil
	})

	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, d := range dirs {
		if entries, _ := fsio.ReadDir(d); len(entries) == 0 {
			_ = fsio.Remove(d)
		}
	}
}
