package storage

// BlockRef represents a content-addressed block of file data.
type BlockRef struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
}

// BlockStatus represents verification result for a block.
type BlockStatus int

const (
	BlockOK BlockStatus = iota
	BlockMissing
	BlockDamaged
)

// BlockCheck holds verification details for a single block.
type BlockCheck struct {
	Hash     string
	Status   BlockStatus
	Files    []string
	Branches []string
}

// FileEntry represents a file split into content-defined blocks.
type FileEntry struct {
	Path   string     `json:"path"`
	Blocks []BlockRef `json:"blocks"`
}

// Equal reports whether two file entries have identical block structures.
func (f *FileEntry) Equal(other *FileEntry) bool {
	if f == nil && other == nil {
		return true
	}
	if f == nil || other == nil {
		return false
	}
	if len(f.Blocks) != len(other.Blocks) {
		return false
	}
	for i := range f.Blocks {
		if f.Blocks[i].Hash != other.Blocks[i].Hash ||
			f.Blocks[i].Size != other.Blocks[i].Size {
			return false
		}
	}
	return true
}

// Fileset represents a snapshot of multiple files.
type Fileset struct {
	ID    string      `json:"id"`
	Files []FileEntry `json:"files"`
}
