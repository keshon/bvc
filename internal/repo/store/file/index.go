package file

import (
	"app/internal/fsio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveIndexReplace overwrites the index completely (for hard resets or clean writes).
func (fc *FileContext) SaveIndexReplace(entries []Entry) error {
	indexPath := filepath.Join(fc.RepoRoot, "index.json")
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	if err := fsio.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		return fmt.Errorf("mkdir index dir: %w", err)
	}
	return fsio.WriteFile(indexPath, data, 0o644)
}

// SaveIndexMerge merges the given entries with any existing index on disk.
// Existing entries with the same path are updated; others are preserved.
func (fc *FileContext) SaveIndexMerge(newEntries []Entry) error {
	existing, _ := fc.LoadIndex() // ignore error if index doesn’t exist

	entryMap := make(map[string]Entry, len(existing)+len(newEntries))

	// Keep old staged entries
	for _, e := range existing {
		entryMap[e.Path] = e
	}

	// Add or overwrite new ones
	for _, e := range newEntries {
		entryMap[e.Path] = e
	}

	// Flatten map
	merged := make([]Entry, 0, len(entryMap))
	for _, e := range entryMap {
		merged = append(merged, e)
	}

	return fc.SaveIndexReplace(merged)
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
