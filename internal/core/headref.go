package core

import (
	"app/internal/config"
	"os"
	"path/filepath"
)

type HeadRef string

func (h HeadRef) String() string {
	return string(h)
}

// GetHeadRef returns the current HEAD ref.
func GetHeadRef() (HeadRef, error) {
	data, err := os.ReadFile(filepath.Join(config.RepoDir, "HEAD"))
	if err != nil {
		return "", err
	}
	ref := string(data)
	return HeadRef(ref[len("ref: "):]), nil
}

// SetHeadRef sets the HEAD ref to the given branch.
func SetHeadRef(branch string) (HeadRef, error) {
	err := os.WriteFile(filepath.Join(config.RepoDir, "HEAD"), []byte("ref: "+branch), 0644)
	if err != nil {
		return "", err
	}
	return GetHeadRef()
}
