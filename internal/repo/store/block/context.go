package block

const (
	minChunkSize = 2 * 1024 * 1024 // 2 MiB
	maxChunkSize = 8 * 1024 * 1024 // 8 MiB
	rollMod      = 4096
)

// BlockRef describes one physical block of content.
type BlockRef struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
}

// BlockStatus indicates the state of a block on disk.
type BlockStatus int

const (
	OK BlockStatus = iota
	Missing
	Damaged
)

type BlockCheck struct {
	Hash     string
	Status   BlockStatus
	Files    []string
	Branches []string
}

// BlockContext handles all object-level storage (.bvc/objects)
type BlockContext struct {
	Root string
}
