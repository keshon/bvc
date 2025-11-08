package block

import (
	"app/internal/fsio"
	"path/filepath"
	"strings"
)

// CleanupTemp removes orphaned temp files from the block storage directory.
func (bc *BlockContext) CleanupTemp() error {
	entries, err := fsio.ReadDir(bc.Root)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "tmp-") || strings.HasPrefix(name, ".tmp-") {
			p := filepath.Join(bc.Root, name)
			if fi, err := fsio.StatFile(p); err != nil || fi.Size() == 0 {
				_ = fsio.Remove(p)
			}
		}
	}
	return nil
}
