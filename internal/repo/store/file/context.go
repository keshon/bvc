package file

import "app/internal/repo/store/block"

// FileContext wraps file-level operations that depend on BlockContext.
type FileContext struct {
	Root     string
	RepoRoot string
	Blocks   *block.BlockContext
}
