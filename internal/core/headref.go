package core

import (
	"fmt"
	"os"
	"path/filepath"

	"app/internal/config"
)

type HeadRef string

func (h HeadRef) String() string {
	return string(h)
}

// GetHeadRef returns the current HEAD ref.
// Returns ref object and error
func GetHeadRef() (HeadRef, error) {
	data, err := os.ReadFile(filepath.Join(config.RepoDir, "HEAD"))
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD: %w", err)
	}

	const prefix = "ref: "
	if len(data) < len(prefix) || string(data[:len(prefix)]) != prefix {
		return "", fmt.Errorf("invalid HEAD content: %q", string(data))
	}

	ref := string(data[len(prefix):])
	return HeadRef(ref), nil
}

// SetHeadRef sets the HEAD ref to the given branch.
// Returns ref object and error
func SetHeadRef(branch string) (HeadRef, error) {
	path := filepath.Join(config.RepoDir, "HEAD")
	if err := os.WriteFile(path, []byte("ref: "+branch), 0o644); err != nil {
		return "", fmt.Errorf("failed to write HEAD: %w", err)
	}

	return HeadRef(branch), nil
}
