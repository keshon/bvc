package core

import (
	"fmt"
	"os"
	"path/filepath"
)

type HeadRef string

func (h HeadRef) String() string { return string(h) }

// GetHeadRef reads HEAD for this repository.
func (r *Repository) GetHeadRef() (HeadRef, error) {
	data, err := os.ReadFile(r.HeadFile)
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD %q: %w", r.HeadFile, err)
	}

	const prefix = "ref: "
	if len(data) < len(prefix) || string(data[:len(prefix)]) != prefix {
		// Allow detached HEAD as raw value (optional). For now we require ref: ...
		return "", fmt.Errorf("invalid HEAD content: %q", string(data))
	}

	ref := string(data[len(prefix):])
	return HeadRef(ref), nil
}

// SetHeadRef sets HEAD to the given branch reference (e.g. "branches/main").
// Accepts either "branches/<name>" or just "<name>" (interpreted as branch name).
func (r *Repository) SetHeadRef(branch string) (HeadRef, error) {
	// normalize: if branch doesn't contain '/', treat as branch name
	refVal := branch
	if filepath.Base(branch) == branch { // no slash present
		refVal = "branches/" + branch
	}
	content := "ref: " + refVal
	if err := os.WriteFile(r.HeadFile, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write HEAD %q: %w", r.HeadFile, err)
	}
	return HeadRef(refVal), nil
}
