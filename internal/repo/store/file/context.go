package file

import "app/internal/repo/store/block"

// FileContext manages file-level operations (staging, restore, scan) with abstracted dependencies.
type FileContext struct {
	Root     string
	RepoRoot string
	FS       FS
	Blocks   BlockStore
}

// Entry represents a tracked file and its content blocks.
type Entry struct {
	Path   string
	Blocks []block.BlockRef
}

// Equal compares two entries by their blocks.
func (e *Entry) Equal(other *Entry) bool {
	if e == nil && other == nil {
		return true
	}
	if e == nil || other == nil {
		return false
	}
	if len(e.Blocks) != len(other.Blocks) {
		return false
	}
	for i := range e.Blocks {
		if e.Blocks[i].Hash != other.Blocks[i].Hash || e.Blocks[i].Size != other.Blocks[i].Size {
			return false
		}
	}
	return true
}
