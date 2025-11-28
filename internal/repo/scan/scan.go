package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/ignore"
	"github.com/keshon/bvc/internal/repo/meta"
	"github.com/keshon/bvc/internal/util"
)

type ScanResult struct {
	Tracked   []string // files from HEAD commit
	Untracked []string // files not tracked by HEAD commit
	Staged    []string // files from index
	Ignored   []string // files ignored by ignore rules
}

// Scanner scans repository working tree, index and committed snapshots.
type Scanner struct {
	WorkingTreeDir string
	FS             fs.FS
	Meta           meta.MetaContextInterface
}

// NewScanner creates a scanner. If fs is nil, uses OS FS.
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

// ScanAll walks the working tree and returns three slices:
//   - tracked: files tracked by HEAD commit,
//   - untracked: files not tracked by HEAD commit,
//   - staged: files present in index.json,
//   - ignored: files matched by ignore rules.
//
// Returned slices contain ABSOLUTE paths (joined with WorkingTreeDir).
func (s *Scanner) ScanAll() (*ScanResult, error) {
	result := &ScanResult{}

	repoDir := s.Meta.GetConfig().RepoDir
	repoAbs := filepath.Join(s.WorkingTreeDir, repoDir)

	// skip current running binary
	binBase := filepath.Base(os.Args[0])
	binName := strings.TrimSuffix(binBase, filepath.Ext(binBase))

	matcher := ignore.NewIgnore(s.WorkingTreeDir, s.FS)

	// load staged files
	indexSet := make(map[string]struct{})
	indexPath := filepath.Join(repoDir, "index.json")
	var raw []struct {
		Path string `json:"path"`
	}
	if err := util.ReadJSON(indexPath, &raw); err == nil {
		for _, e := range raw {
			indexSet[filepath.ToSlash(filepath.Clean(e.Path))] = struct{}{}
		}
	}

	// load committed files from HEAD
	committedSet := make(map[string]struct{})
	branch, err := s.Meta.GetCurrentBranch()
	if err == nil {
		curCommit, err2 := s.Meta.GetLastCommitForBranch(branch.Name)
		for curCommit != nil && err2 == nil {
			snap := filepath.Join(s.Meta.GetConfig().SnapshotsDir(), curCommit.FilesetID+".json")
			var payload struct {
				Files []struct {
					Path string `json:"path"`
				}
			}
			if err3 := util.ReadJSON(snap, &payload); err3 == nil {
				for _, f := range payload.Files {
					rel := filepath.ToSlash(filepath.Clean(f.Path))
					committedSet[rel] = struct{}{}
				}
			}
			if len(curCommit.Parents) > 0 {
				curCommit, err2 = s.Meta.GetCommit(curCommit.Parents[0])
			} else {
				break
			}
		}
	}

	// walk working tree
	var walk func(rel string) error
	walk = func(rel string) error {
		dirAbs := filepath.Join(s.WorkingTreeDir, rel)
		entries, err := s.FS.ReadDir(dirAbs)
		if err != nil {
			return err
		}

		for _, e := range entries {
			childRel := filepath.Join(rel, e.Name())
			childAbs := filepath.Join(s.WorkingTreeDir, childRel)

			info, err := e.Info()
			if err != nil || info == nil {
				continue
			}

			// skip repo dir itself
			if filepath.Clean(childAbs) == filepath.Clean(repoAbs) {
				continue
			}

			// skip running binary
			if !info.IsDir() {
				fileName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
				if fileName == binName {
					continue
				}
			}

			relSlash := filepath.ToSlash(childRel)

			// directory
			if info.IsDir() {
				if matcher.Match(relSlash) {
					result.Ignored = append(result.Ignored, childAbs)
					continue
				}
				if err := walk(childRel); err != nil {
					return err
				}
				continue
			}

			// file
			if matcher.Match(relSlash) {
				result.Ignored = append(result.Ignored, childAbs)
				continue
			}

			if _, ok := indexSet[relSlash]; ok {
				result.Staged = append(result.Staged, childAbs)
				continue
			}

			if _, ok := committedSet[relSlash]; ok {
				result.Tracked = append(result.Tracked, childAbs)
				continue
			}

			// untracked
			result.Untracked = append(result.Untracked, childAbs)
		}
		return nil
	}

	if err := walk(""); err != nil {
		return nil, err
	}

	// sort
	sort.Strings(result.Tracked)
	sort.Strings(result.Staged)
	sort.Strings(result.Ignored)
	sort.Strings(result.Untracked)

	return result, nil
}
