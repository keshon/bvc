package core

import (
	"app/internal/config"
	"os"
	"path/filepath"
)

// HeadRef returns the current HEAD ref.
func HeadRef() (string, error) {
	data, err := os.ReadFile(filepath.Join(config.RepoDir, "HEAD"))
	if err != nil {
		return "", err
	}
	ref := string(data)
	return ref[len("ref: "):], nil
}

// SetHeadRef sets the HEAD ref to the given branch.
func SetHeadRef(branch string) error {
	return os.WriteFile(filepath.Join(config.RepoDir, "HEAD"), []byte("ref: "+branch), 0644)
}
