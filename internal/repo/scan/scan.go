package scan

import (
	"path/filepath"
	"sort"

	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/ignore"
	"github.com/keshon/bvc/internal/repo/meta"
	"github.com/keshon/bvc/internal/util"
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

func (s *Scanner) ScanAll() (*ScanResult, error) {
	result := &ScanResult{}

	repoDir := s.Meta.GetConfig().RepoDir
	repoRel := filepath.ToSlash(filepath.Clean(repoDir))

	matcher := ignore.NewIgnore(s.WorkingTreeDir, s.FS)

	// staged from index.json
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

	// committed files
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
			childRel := filepath.ToSlash(filepath.Join(rel, e.Name()))

			info, err := e.Info()
			if err != nil || info == nil {
				continue
			}

			// skip repo metadata directory
			if filepath.ToSlash(childRel) == repoRel {
				continue
			}

			// directories
			if info.IsDir() {
				if matcher.Match(childRel) {
					result.Ignored = append(result.Ignored, childRel)
					continue
				}
				if err := walk(childRel); err != nil {
					return err
				}
				continue
			}

			// ignored files
			if matcher.Match(childRel) {
				result.Ignored = append(result.Ignored, childRel)
				continue
			}

			// staged
			if _, ok := indexSet[childRel]; ok {
				result.Staged = append(result.Staged, childRel)
				continue
			}

			// tracked
			if _, ok := committedSet[childRel]; ok {
				result.Tracked = append(result.Tracked, childRel)
				continue
			}

			// untracked
			result.Untracked = append(result.Untracked, childRel)
		}
		return nil
	}

	if err := walk(""); err != nil {
		return nil, err
	}

	// sort output
	sort.Strings(result.Tracked)
	sort.Strings(result.Staged)
	sort.Strings(result.Ignored)
	sort.Strings(result.Untracked)

	return result, nil
}
