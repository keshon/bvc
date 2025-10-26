package storage

import (
	"app/internal/config"
	"app/internal/progress"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RestoreFileset reconstructs all files in the snapshot.
func RestoreFileset(fs Fileset, label string) error {
	exe := filepath.Base(os.Args[0])
	bar := progress.NewProgress(len(fs.Files), fmt.Sprintf("Restoring files %s", label))
	defer bar.Finish()

	valid := make(map[string]bool, len(fs.Files))
	for i, entry := range fs.Files {
		clean := filepath.Clean(entry.Path)
		valid[clean] = true

		if filepath.Base(clean) == exe {
			bar.SetCurrent(i + 1)
			continue
		}

		if err := restoreFile(entry); err != nil {
			fmt.Printf("\nWarning: %v\n", err)
		}
		bar.SetCurrent(i + 1)
	}

	cleanupExtraFiles(valid, exe)
	return nil
}

// restoreFile rebuilds a single file from blocks atomically.
func restoreFile(entry FileEntry) error {
	if err := os.MkdirAll(filepath.Dir(entry.Path), 0o755); err != nil {
		return err
	}

	dir := filepath.Dir(entry.Path)
	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	writer := bufio.NewWriterSize(tmpFile, 256*1024)

	for _, block := range entry.Blocks {
		data, err := readBlock(block.Hash)
		if err != nil {
			tmpFile.Close()
			return fmt.Errorf("missing block %s for file %s", block.Hash, entry.Path)
		}
		if _, err := writer.Write(data); err != nil {
			tmpFile.Close()
			return fmt.Errorf("write failed for %s: %w", entry.Path, err)
		}
	}
	if err := writer.Flush(); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, entry.Path)
}

// cleanupExtraFiles removes untracked files and empty directories.
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

	if err := removeEmptyDirs("."); err != nil {
		fmt.Printf("\nWarning: cleanup failed: %v\n", err)
	}
}

// removeEmptyDirs recursively deletes empty directories.
func removeEmptyDirs(root string) error {
	var dirs []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path == root || path == config.RepoDir {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}
