package file

import (
	"app/internal/fsio"
	"app/internal/progress"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Restore rebuilds files from entries (e.g., from a snapshot).
func (fc *FileContext) RestoreFilesToWorkingTree(entries []Entry, label string) error {
	if fc.Blocks == nil {
		return fmt.Errorf("no BlockManager attached")
	}

	exe := filepath.Base(os.Args[0])
	bar := progress.NewProgress(len(entries), fmt.Sprintf("Restoring %s", label))
	defer bar.Finish()

	// Build valid file map from Fileset entries
	valid := make(map[string]bool, len(entries))
	for _, e := range entries {
		valid[filepath.Clean(e.Path)] = true
	}

	// Include staged files so we don't delete them
	staged, err := fc.LoadIndex()
	if err != nil {
		fmt.Printf("\nWarning: %v\n", err)
	}
	for _, s := range staged {
		valid[filepath.Clean(s.Path)] = true
	}

	// Restore Fileset entries first
	for _, e := range entries {
		if filepath.Base(e.Path) == exe {
			continue
		}
		if err := fc.restoreFile(e); err != nil {
			fmt.Printf("\nWarning: %v\n", err)
		}
		bar.Increment()
	}

	// Now prune untracked files safely
	fc.removeUntracked(valid, exe) // TODO: check this method - it's probably broken
	return nil
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

// TODO: its probably broken (fc.RepoRoot should be meaningless here)
func (fc *FileContext) removeUntracked(valid map[string]bool, exe string) {
	matcher := NewIgnore()

	var dirs []string
	filepath.WalkDir(fc.RepoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}

		clean := filepath.Clean(path)

		// Skip ignored files and dirs
		if matcher.Match(clean) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			dirs = append(dirs, clean)
			return nil
		}

		if valid[clean] || filepath.Base(clean) == exe {
			return nil
		}

		_ = fsio.Remove(clean)
		return nil
	})

	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, d := range dirs {
		if entries, _ := fsio.ReadDir(d); len(entries) == 0 {
			_ = fsio.Remove(d)
		}
	}
}
