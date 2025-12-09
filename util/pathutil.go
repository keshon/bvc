package util

import (
	"path"
	"path/filepath"
	"strings"
)

func Normalize(p string) string {
	p = filepath.Clean(p)
	p = strings.ReplaceAll(p, "\\", "/")
	return p
}

func JoinNorm(elem ...string) string {
	return Normalize(path.Join(elem...))
}
