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

func NewIgnore() *Ignore {
	m := &Ignore{static: make(map[string]bool)}

	// Add default ignored files
	for _, s := range config.DefaultIgnoredFiles {
		m.static[filepath.Clean(s)] = true
	}

	// Try load .bvc-ignore
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

func (m *Ignore) Match(path string) bool {
	clean := filepath.Clean(path)

	// Direct static match
	if m.static[clean] {
		return true
	}

	// Directory or pattern match
	base := filepath.Base(clean)
	for _, pat := range m.pattern {
		if matched, _ := filepath.Match(pat, base); matched {
			return true
		}
		if strings.HasPrefix(clean, pat) {
			return true
		}
	}

	return false
}
