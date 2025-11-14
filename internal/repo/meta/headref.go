package meta

import (
	"fmt"
	"path/filepath"
)

type HeadRef string

func (h HeadRef) String() string { return string(h) }

// GetHeadRef reads HEAD for this repository.
func (mc *MetaContext) GetHeadRef() (HeadRef, error) {
	data, err := mc.FS.ReadFile(mc.Config.HeadFile())
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD %q: %w", mc.Config.HeadFile(), err)
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
func (mc *MetaContext) SetHeadRef(branch string) (HeadRef, error) {
	// normalize: if branch doesn't contain '/', treat as branch name
	refVal := branch
	if filepath.Base(branch) == branch { // no slash present
		refVal = "branches/" + branch
	}
	content := "ref: " + refVal
	if err := mc.FS.WriteFile(mc.Config.HeadFile(), []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write HEAD %q: %w", mc.Config.HeadFile(), err)
	}
	return HeadRef(refVal), nil
}
