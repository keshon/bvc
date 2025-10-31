package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// StageFiles writes staged entries (index) to disk.
func (fm *FileManager) StageFiles(entries []Entry) error {
	indexPath := filepath.Join(fm.Root, "index.json")
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		return fmt.Errorf("mkdir index dir: %w", err)
	}
	return os.WriteFile(indexPath, data, 0644)
}

// ClearIndex removes the staging index.
func (fm *FileManager) ClearIndex() error {
	indexPath := filepath.Join(fm.Root, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(indexPath)
}

// GetIndexFiles loads staged entries from disk.
func (fm *FileManager) GetIndexFiles() ([]Entry, error) {
	indexPath := filepath.Join(fm.Root, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, nil
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal index: %w", err)
	}
	return entries, nil
}
