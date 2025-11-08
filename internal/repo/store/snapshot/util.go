package snapshot

import (
	"app/internal/repo/store/file"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/zeebo/xxh3"
)

// HashFileset generates a stable hash for a given fileset’s contents.
func HashFileset(entries []file.Entry) string {
	paths := make([]string, 0, len(entries))
	index := make(map[string]file.Entry, len(entries))
	for _, f := range entries {
		clean := filepath.Clean(f.Path)
		paths = append(paths, clean)
		index[clean] = f
	}
	sort.Strings(paths)

	data := make([]byte, 0, len(paths)*64)
	for _, p := range paths {
		for _, b := range index[p].Blocks {
			data = append(data, []byte(b.Hash+"\n")...)
		}
	}

	return fmt.Sprintf("%x", xxh3.Hash128(data).Bytes())
}
