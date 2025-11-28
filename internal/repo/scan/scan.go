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
//   - tracked: files tracked by HEAD commit (but not staged),
//   - staged: files present in index.json,
//   - ignored: files matched by ignore rules.
//
// Returned slices contain ABSOLUTE paths (joined with WorkingTreeDir).
func (s *Scanner) ScanAll() (tracked, staged, ignored []string, err error) {
	repoDir := s.Meta.GetConfig().RepoDir
	repoAbs := filepath.Join(s.WorkingTreeDir, repoDir)

	// Skip current running binary
	binBase := filepath.Base(os.Args[0])
	binName := strings.TrimSuffix(binBase, filepath.Ext(binBase))

	matcher := ignore.NewIgnore(s.WorkingTreeDir, s.FS)

	// Load staged files (index.json)
	indexSet := make(map[string]struct{})
	{
		indexPath := filepath.Join(repoDir, "index.json")
		var raw []struct {
			Path string `json:"path"`
		}
		if err := util.ReadJSON(indexPath, &raw); err == nil {
			for _, e := range raw {
				indexSet[filepath.ToSlash(filepath.Clean(e.Path))] = struct{}{}
			}
		}
	}

	// Load committed files from HEAD only
	committedSet := make(map[string]struct{})
	{
		branch, err := s.Meta.GetCurrentBranch()
		if err == nil {
			lastCommit, err := s.Meta.GetLastCommitForBranch(branch.Name)
			if err == nil && lastCommit != nil && lastCommit.FilesetID != "" {
				snap := filepath.Join(s.Meta.GetConfig().SnapshotsDir(), lastCommit.FilesetID+".json")

				var payload struct {
					Files []struct {
						Path string `json:"path"`
					} `json:"files"`
				}
				if err := util.ReadJSON(snap, &payload); err == nil {
					for _, f := range payload.Files {
						rel := filepath.ToSlash(filepath.Clean(f.Path))
						committedSet[rel] = struct{}{}
					}
				}
			}
		}
	}

	// Walk working tree
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

			// Directory
			if info.IsDir() {
				if matcher.Match(relSlash) {
					ignored = append(ignored, childAbs)
					continue
				}
				if err := walk(childRel); err != nil {
					return err
				}
				continue
			}

			// File
			if matcher.Match(relSlash) {
				ignored = append(ignored, childAbs)
				continue
			}

			if _, ok := indexSet[relSlash]; ok {
				staged = append(staged, childAbs)
				continue
			}

			if _, ok := committedSet[relSlash]; ok {
				tracked = append(tracked, childAbs)
				continue
			}

			// untracked - ignore for tracked files (caller can handle separately)
		}

		return nil
	}

	if err := walk(""); err != nil {
		return nil, nil, nil, err
	}

	sort.Strings(tracked)
	sort.Strings(staged)
	sort.Strings(ignored)

	return tracked, staged, ignored, nil
}
