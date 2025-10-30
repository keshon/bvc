package file

import (
	"app/internal/config"
	"app/internal/progress"
	"app/internal/storage/block"
	"app/internal/util"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RestoreFiles rebuilds all files from their entries in a snapshot.
func RestoreFiles(entries []Entry, label string) error {
	exe := filepath.Base(os.Args[0])
	bar := progress.NewProgress(len(entries), fmt.Sprintf("Restoring %s", label))
	defer bar.Finish()

	// Create valid map sequentially (no concurrency)
	valid := make(map[string]bool, len(entries))
	for _, e := range entries {
		valid[filepath.Clean(e.Path)] = true
	}

	// Restore files in parallel
	err := util.Parallel(entries, util.WorkerCount()*2, func(e Entry) error {
		if filepath.Base(e.Path) == exe {
			// Skip restoring executable
			bar.Increment()
			return nil
		}

		if err := restoreFile(e); err != nil {
			fmt.Printf("\nWarning: %v\n", err)
		}

		bar.Increment()
		return nil
	})

	// Remove untracked files
	pruneUntrackedFiles(valid, exe)

	return err
}

func restoreFile(e Entry) error {
	if err := os.MkdirAll(filepath.Dir(e.Path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(e.Path), "tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	writer := bufio.NewWriterSize(tmp, 4*1024*1024)
	for _, b := range e.Blocks {
		data, err := block.ReadBlock(b.Hash)
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

func pruneUntrackedFiles(valid map[string]bool, exe string) {
	var dirs []string
	filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			if path == "." || path == config.RepoDir || strings.HasPrefix(path, config.RepoDir+string(os.PathSeparator)) {
				return filepath.SkipDir
			}
			dirs = append(dirs, path)
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
