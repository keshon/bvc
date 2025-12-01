package scan

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/ignore"
	"github.com/keshon/bvc/internal/repo/meta"
)

type ScanResult struct {
	Tracked   []string // files from HEAD commit (repo-relative)
	Untracked []string // files not tracked by HEAD commit (repo-relative)
	Staged    []string // files from index (repo-relative)
	Ignored   []string // files ignored by ignore rules (repo-relative)
}

type Scanner struct {
	WorkingTreeDir string
	FS             fs.FS
	Meta           meta.MetaContextInterface
}

func NewScanner(workingTreeDir string, mc meta.MetaContextInterface, fsys fs.FS) *Scanner {
	if fsys == nil {
		fsys = fs.NewOSFS()
	}
	return &Scanner{
		WorkingTreeDir: workingTreeDir,
		FS:             fsys,
		Meta:           mc,
	}
}

func (s *Scanner) readJSON(path string, v interface{}) error {
	data, err := s.FS.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *Scanner) ScanAll() (*ScanResult, error) {
	r := &ScanResult{}
	cfg := s.Meta.GetConfig()

	// repo dir = wt/.bvc
	repoDir := filepath.ToSlash(filepath.Clean(cfg.RepoDir))

	// ignore
	ign := ignore.NewIgnore(s.WorkingTreeDir, s.FS)

	// load staged
	indexSet := map[string]struct{}{}
	indexPath := filepath.ToSlash(filepath.Join(repoDir, "index.json"))

	var rawIdx []struct {
		Path string `json:"path"`
	}
	if err := s.readJSON(indexPath, &rawIdx); err == nil {
		for _, e := range rawIdx {
			indexSet[cleanRel(e.Path)] = struct{}{}
		}
	}

	// load committed
	committedSet := map[string]struct{}{}
	branch, err := s.Meta.GetCurrentBranch()
	if err == nil {
		last, err := s.Meta.GetLastCommitForBranch(branch.Name)
		for last != nil && err == nil {
			snapPath := filepath.ToSlash(filepath.Join(cfg.SnapshotsDir(), last.FilesetID+".json"))

			var payload struct {
				Files []struct {
					Path string `json:"path"`
				}
			}

			if err2 := s.readJSON(snapPath, &payload); err2 == nil {
				for _, f := range payload.Files {
					committedSet[cleanRel(f.Path)] = struct{}{}
				}
			}

			if len(last.Parents) == 0 {
				break
			}
			last, err = s.Meta.GetCommit(last.Parents[0])
		}
	}

	// walk filesystem
	wt := filepath.ToSlash(filepath.Clean(s.WorkingTreeDir))

	var walk func(rel string) error
	walk = func(rel string) error {
		abs := filepath.Join(wt, rel)

		entries, err := s.FS.ReadDir(abs)
		if err != nil {
			return err
		}

		for _, e := range entries {
			name := e.Name()
			childRel := filepath.ToSlash(filepath.Join(rel, name))

			// skip .bvc fully
			if isRepoDir(childRel, repoDir) {
				continue
			}

			fi, err := e.Info()
			if err != nil {
				continue
			}

			// directory recursively
			if fi.IsDir() {
				if ign.Match(childRel) {
					r.Ignored = append(r.Ignored, childRel)
					continue
				}
				if err := walk(childRel); err != nil {
					return err
				}
				continue
			}

			// ignored file
			if ign.Match(childRel) {
				r.Ignored = append(r.Ignored, childRel)
				continue
			}

			// staged
			if _, ok := indexSet[childRel]; ok {
				r.Staged = append(r.Staged, childRel)
				continue
			}

			// tracked (commit)
			if _, ok := committedSet[childRel]; ok {
				r.Tracked = append(r.Tracked, childRel)
				continue
			}

			// untracked
			r.Untracked = append(r.Untracked, childRel)
		}

		return nil
	}

	if err := walk(""); err != nil {
		return nil, err
	}

	sort.Strings(r.Tracked)
	sort.Strings(r.Staged)
	sort.Strings(r.Ignored)
	sort.Strings(r.Untracked)

	return r, nil
}

func cleanRel(p string) string {
	return filepath.ToSlash(filepath.Clean(p))
}

// correct detection of .bvc folder relative to working tree
func isRepoDir(childRel, repoDir string) bool {
	// repoDir may be "wt/.bvc"
	// repoDir basename is ".bvc"
	return strings.HasPrefix(childRel, filepath.Base(repoDir))
}
