package workfs

import (
	"bufio"
	"bvc/util"
	"os"
	"path/filepath"
	"strings"
)

type FileIgnore struct {
	static   map[string]bool
	patterns []string
}

func LoadIgnore(root string) *FileIgnore {
	m := &FileIgnore{static: map[string]bool{}}
	m.static[DefaultRepoDir] = true
	p := filepath.Join(root, ".bvc-ignore")
	f, err := os.Open(p)
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m.patterns = append(m.patterns, line)
	}
	return m
}

func (ig *FileIgnore) Match(rel string, isDir bool) bool {
	rel = util.Normalize(rel)
	rel = filepath.ToSlash(rel)
	if ig.static[rel] {
		return true
	}
	for _, p := range ig.patterns {
		p = util.Normalize(p)
		p = filepath.ToSlash(p)
		if strings.HasSuffix(p, "/") && isDir {
			p = strings.TrimSuffix(p, "/")
			if strings.HasPrefix(rel, p) {
				return true
			}
		}
		ok, _ := filepath.Match(p, rel)
		if ok {
			return true
		}
	}
	return false
}
