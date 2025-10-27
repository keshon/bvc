package file

import (
	"encoding/json"
	"os"
	"path/filepath"

	"app/internal/config"
)

var indexFile = filepath.Join(config.RepoDir, "index.json")

// Stage Files writes files to the index (staging area)
func StageFiles(entries []Entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(indexFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(indexFile, data, 0644)
}

// ClearIndex clears the staging area
func ClearIndex() error {
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(indexFile)
}

// GetIndexFiles reads the staging area
func GetIndexFiles() ([]Entry, error) {
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return nil, nil
	}
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
