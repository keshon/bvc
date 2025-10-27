package block

type BlockRef struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
}

type BlockStatus int

const (
	OK BlockStatus = iota
	Missing
	Damaged
)

type BlockCheck struct {
	Hash        string
	BlockStatus BlockStatus
	Files       []string
	Branches    []string
}
