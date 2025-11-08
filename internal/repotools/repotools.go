package repotools

// BlockInfo holds metadata about a block in the repository
type BlockInfo struct {
	Size     int64
	Files    map[string]struct{}
	Branches map[string]struct{}
}
