package workfs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/keshon/bvc/pkg/util"
)

const DefaultRepoDir = ".bvc"

type Ignore interface {
	Match(rel string, isDir bool) bool
}

// WalkFiles walks project files and calls fn for regular files (skips repo dir and ignored).
func WalkFiles(root string, ig Ignore, fn func(rel string, f *os.File) error) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relRaw, _ := filepath.Rel(root, path)
		rel := util.Normalize(relRaw)

		if filepath.Base(path) == DefaultRepoDir || ig.Match(rel, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			return fn(rel, f)
		}
		return nil
	})
}

// RestoreFile writes blocks (using provided getter) into dst via tmp + rename.
func RestoreFile(root, rel string, blocks []string, getBlock func(hash string, w io.Writer) error) error {
	rel = util.Normalize(rel)
	dst := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".bvc.tmp"
	tf, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer func() {
		tf.Close()
		_ = os.Remove(tmp) // best-effort cleanup
	}()
	for _, h := range blocks {
		if err := getBlock(h, tf); err != nil {
			return err
		}
	}
	if err := tf.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
