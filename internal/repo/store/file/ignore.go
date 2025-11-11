package file

import (
	"app/internal/config"
	"app/internal/fsio"
	"bufio"
	"path/filepath"
	"strings"
)

type Ignore struct {
	static  map[string]bool
	pattern []string
}

// NewIgnore loads defaults and .bvc-ignore
func NewIgnore() *Ignore {
	m := &Ignore{static: make(map[string]bool)}

	// Default ignored files
	for _, s := range config.DefaultIgnoredFiles {
		m.static[filepath.Clean(s)] = true
	}

	// Load .bvc-ignore
	f, err := fsio.Open(config.IgnoredFilesFile)
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
	return matchSegments(strings.Split(pattern, "/"), strings.Split(path, "/"))
}

// matchSegments matches pattern segments recursively
func matchSegments(pats, parts []string) bool {
	for len(pats) > 0 {
		p := pats[0]
		pats = pats[1:]

		if p == "**" {
			if len(pats) == 0 {
				return true // trailing ** matches anything
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

	return len(parts) == 0
}
