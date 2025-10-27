package file

import (
	"app/internal/config"
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

	cleanupExtraFiles(valid, exe)
	return nil
}

// restoreSingle rebuilds a single file from its blocks atomically.
func restoreSingle(entry Entry) error {
	if err := os.MkdirAll(filepath.Dir(entry.Path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(entry.Path), "tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	writer := bufio.NewWriterSize(tmp, 256*1024)
	for _, ref := range entry.Blocks {
		data, err := block.Read(ref.Hash)
		if err != nil {
			tmp.Close()
			return fmt.Errorf("missing block %s for file %s", ref.Hash, entry.Path)
		}
		if _, err := writer.Write(data); err != nil {
			tmp.Close()
			return fmt.Errorf("write failed for %s: %w", entry.Path, err)
		}
	}
	if err := writer.Flush(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	return os.Rename(tmpPath, entry.Path)
}

// cleanupExtraFiles removes files not in the snapshot and empties dirs.
func cleanupExtraFiles(valid map[string]bool, exe string) {
	filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || path == config.RepoDir {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		clean := filepath.Clean(path)
		if filepath.Base(clean) != exe && !valid[clean] {
			_ = os.Remove(path)
		}
		return nil
	})
	removeEmptyDirs(".")
}

func removeEmptyDirs(root string) {
	var dirs []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() || path == root || path == config.RepoDir {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, dir := range dirs {
		if entries, _ := os.ReadDir(dir); len(entries) == 0 {
			_ = os.Remove(dir)
		}
	}
}
