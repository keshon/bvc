package block

import (
	"app/internal/config"
	"os"
	"path/filepath"
)

// CleanupTmp removes orphaned temp files.
func CleanupTmp() error {
	entries, err := os.ReadDir(config.ObjectsDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) > 4 && name[:4] == "tmp-" {
			p := filepath.Join(config.ObjectsDir, name)
			if fi, err := os.Stat(p); err != nil || fi.Size() == 0 {
				_ = os.Remove(p)
			}
		}
	}
	return nil
}
