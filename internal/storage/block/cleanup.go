package block

import (
	"app/internal/config"
	"os"
	"path/filepath"
)

func CleanupTmp() error {
	entries, err := os.ReadDir(config.ObjectsDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 4 && e.Name()[:4] == "tmp-" {
			p := filepath.Join(config.ObjectsDir, e.Name())
			if fi, err := os.Stat(p); err != nil || fi.Size() == 0 {
				_ = os.Remove(p)
			}
		}
	}
	return nil
}
