package file

import (
	"app/internal/progress"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// Restore rebuilds files from entries (e.g., from a snapshot).
func (fc *FileContext) RestoreFilesToWorkingTree(entries []Entry, label string) error {
	if fc.Blocks == nil {
		return fmt.Errorf("no BlockContext attached")
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
		return fmt.Errorf("failed to load index: %w", err) //fmt.Printf("\nWarning: %v\n", err)
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
			return fmt.Errorf("restoring file %s: %w", e.Path, err) //fmt.Printf("\nWarning: %v\n", err)
		}
		bar.Increment()
	}

	// Now prune untracked files safely
	return nil
}

func (fc *FileContext) restoreFile(e Entry) error {
	if err := fc.FS.MkdirAll(filepath.Dir(e.Path), 0o755); err != nil {
		return err
	}

	tmp, err := fc.FS.CreateTempFile(filepath.Dir(e.Path), "tmp-*")
	if err != nil {
		return err
	}
	defer fc.FS.Remove(tmp.Name())
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

	return fc.FS.Rename(tmp.Name(), e.Path)
}
