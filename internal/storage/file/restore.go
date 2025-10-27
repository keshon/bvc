package file

import (
	"app/internal/progress"
	"app/internal/storage/block"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RestoreAll rebuilds all files from their entries in a snapshot.
func RestoreAll(entries []Entry, label string) error {
	exe := filepath.Base(os.Args[0])
	bar := progress.NewProgress(len(entries), fmt.Sprintf("Restoring %s", label))
	defer bar.Finish()

	valid := make(map[string]bool, len(entries))
	for i, entry := range entries {
		clean := filepath.Clean(entry.Path)
		valid[clean] = true

		if filepath.Base(clean) == exe {
			bar.SetCurrent(i + 1)
			continue
		}

		if err := restoreSingle(entry); err != nil {
			fmt.Printf("\nWarning: %v\n", err)
		}
		bar.SetCurrent(i + 1)
	}

	cleanupExtra(valid, exe)
	return nil
}

func restoreSingle(e Entry) error {
	if err := os.MkdirAll(filepath.Dir(e.Path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(e.Path), "tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	writer := bufio.NewWriterSize(tmp, 256*1024)
	for _, b := range e.Blocks {
		data, err := block.Read(b.Hash)
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

	return os.Rename(tmp.Name(), e.Path)
}

func cleanupExtra(valid map[string]bool, exe string) {
	var dirs []string
	filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			if path != "." {
				dirs = append(dirs, path)
			}
			return nil
		}
		if !valid[filepath.Clean(path)] && filepath.Base(path) != exe {
			_ = os.Remove(path)
		}
		return nil
	})

	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, d := range dirs {
		if entries, _ := os.ReadDir(d); len(entries) == 0 {
			_ = os.Remove(d)
		}
	}
}
