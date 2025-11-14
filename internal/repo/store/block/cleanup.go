package block

import (
	"path/filepath"
	"strings"
)

// CleanupTemp removes orphaned temp files from the block storage directory.
func (bc *BlockContext) CleanupTemp() error {
	entries, err := bc.FS.ReadDir(bc.Root)
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
			if fi, err := bc.FS.Stat(p); err != nil || fi.Size() == 0 {
				_ = bc.FS.Remove(p)
			}
		}
	}
	return nil
}
