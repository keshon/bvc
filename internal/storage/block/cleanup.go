package block

import (
	"os"
	"path/filepath"
	"strings"
)

// CleanupTemp removes orphaned temp files from the block storage directory.
func (bm *BlockManager) CleanupTemp() error {
	entries, err := os.ReadDir(bm.Root)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "tmp-") || strings.HasPrefix(name, ".tmp-") {
			p := filepath.Join(bm.Root, name)
			if fi, err := os.Stat(p); err != nil || fi.Size() == 0 {
				_ = os.Remove(p)
			}
		}
	}
	return nil
}
