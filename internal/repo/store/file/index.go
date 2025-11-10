package file

import (
	"app/internal/fsio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveIndex writes staged entries (index) to disk.
func (fc *FileContext) SaveIndex(entries []Entry) error {
	indexPath := filepath.Join(fc.RepoRoot, "index.json")
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	if err := fsio.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		return fmt.Errorf("mkdir index dir: %w", err)
	}
	return fsio.WriteFile(indexPath, data, 0644)
}

// ClearIndex removes the staging index.
func (fc *FileContext) ClearIndex() error {
	indexPath := filepath.Join(fc.RepoRoot, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil
	}
	return fsio.Remove(indexPath)
}

// LoadIndex loads staged entries from disk.
func (fc *FileContext) LoadIndex() ([]Entry, error) {
	indexPath := filepath.Join(fc.RepoRoot, "index.json")
	if _, err := fsio.StatFile(indexPath); fsio.IsNotExist(err) {
		return nil, nil
	}
	data, err := fsio.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal index: %w", err)
	}
	return entries, nil
}
