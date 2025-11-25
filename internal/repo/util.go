package repo

import (
	"path/filepath"
	"strings"

	"github.com/keshon/bvc/internal/repo/file"
)

func filterEntriesByPatterns(entries []file.Entry, patterns []string) []file.Entry {
	var out []file.Entry

	for _, p := range patterns {
		if p == "." {
			return entries
		}
		for _, e := range entries {
			ok, _ := filepath.Match(p, filepath.Base(e.Path))
			if ok || strings.HasPrefix(e.Path, p) {
				out = append(out, e)
			}
		}
	}

	return out
}
