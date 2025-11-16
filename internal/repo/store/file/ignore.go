package file

import (
	"bufio"
	"path/filepath"
	"strings"

	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"
)

type Ignore struct {
	static  map[string]bool
	pattern []string
}

// NewIgnore loads defaults and .bvc-ignore from the given repo root.
func NewIgnore(repoRoot string, fs fs.FS) *Ignore {
	m := &Ignore{static: make(map[string]bool)}

	// Default ignored files
	for _, s := range config.DefaultIgnoredFiles {
		m.static[filepath.Clean(s)] = true
	}

	ignorePath := filepath.Join(repoRoot, config.IgnoredFilesFile)
	f, err := fs.Open(ignorePath)
	if err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			m.pattern = append(m.pattern, line)
		}
		f.Close()
	}
	return m
}

// Match returns true if the path should be ignored
// supply only relative paths
func (m *Ignore) Match(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))

	// static exact match
	if m.static[clean] {
		return true
	}

	// pattern match
	for _, pat := range m.pattern {
		if matchPattern(pat, clean) {
			return true
		}
	}

	return false
}

// matchPattern handles *, ?, and ** like Git
func matchPattern(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Special case: both empty => match
	if pattern == "" && path == "" {
		return true
	}

	return matchSegments(strings.Split(pattern, "/"), strings.Split(path, "/"))
}

// matchSegments matches pattern segments recursively
func matchSegments(pats, parts []string) bool {
	// remove trailing empty segments from pattern (e.g., "dir/**/" → ["dir", "**"])
	for len(pats) > 0 && pats[len(pats)-1] == "" {
		pats = pats[:len(pats)-1]
	}

	for len(pats) > 0 {
		p := pats[0]
		pats = pats[1:]

		if p == "**" {
			// ** matches any number of path segments (including zero)
			if len(pats) == 0 {
				return true
			}
			for i := 0; i <= len(parts); i++ {
				if matchSegments(pats, parts[i:]) {
					return true
				}
			}
			return false
		}

		if len(parts) == 0 {
			return false
		}

		ok, _ := filepath.Match(p, parts[0])
		if !ok {
			return false
		}

		parts = parts[1:]
	}

	// if we've consumed all patterns, match succeeds if no remaining parts
	return len(parts) == 0
}
