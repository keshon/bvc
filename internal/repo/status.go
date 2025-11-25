// repo/status.go
package repo

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/keshon/bvc/internal/repo/file"
)

type FileStatus struct {
	Path     string
	Staged   string
	Unstaged string
}

type RepoStatus struct {
	Branch    string
	Items     []FileStatus
	Untracked []string
	Ignored   []string
	IsClean   bool
}

type StatusOptions struct {
	UntrackedMode string // "no", "normal", "all"
	ShowIgnored   bool
}

func (r *Repository) Status(opts StatusOptions) (*RepoStatus, error) {
	if opts.UntrackedMode == "" {
		opts.UntrackedMode = "normal"
	}

	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("no current branch: %w", err)
	}

	// HEAD
	headFiles := map[string]file.Entry{}
	if commitID, _ := r.Meta.GetLastCommitID(branch.Name); commitID != "" {
		fs, err := r.GetCommittedFileset(commitID)
		if err != nil {
			return nil, err
		}
		for _, e := range fs.Files {
			headFiles[filepath.Clean(e.Path)] = e
		}
	}

	// Working tree, staged, ignored
	trackedFS, stagedFS, ignoredFS, err := r.Snapshot.BuildAllRepositoryFilesets()
	if err != nil {
		return nil, fmt.Errorf("scan working tree: %w", err)
	}

	tracked := make(map[string]file.Entry)
	for _, e := range trackedFS.Files {
		tracked[filepath.Clean(e.Path)] = e
	}

	staged := make(map[string]file.Entry)
	for _, e := range stagedFS.Files {
		staged[filepath.Clean(e.Path)] = e
	}

	ignored := make(map[string]file.Entry)
	for _, e := range ignoredFS.Files {
		ignored[filepath.Clean(e.Path)] = e
	}

	// Collect all relevant file paths
	allPaths := make(map[string]struct{})
	for p := range headFiles {
		allPaths[p] = struct{}{}
	}
	for p := range staged {
		allPaths[p] = struct{}{}
	}
	for p := range tracked {
		allPaths[p] = struct{}{}
	}

	var paths []string
	for p := range allPaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	status := RepoStatus{
		Branch: branch.Name,
	}

	for _, p := range paths {
		h, inHead := headFiles[p]
		s, inStaged := staged[p]
		w, inWork := tracked[p]

		var stg, ust string

		// staged
		switch {
		case inStaged && !inHead:
			stg = "A"
		case inStaged && inHead && !h.Equal(&s):
			stg = "M"
		case inHead && !inWork && inStaged:
			stg = "D"
		}

		// unstaged
		switch {
		case inWork && inHead && !h.Equal(&w):
			ust = "M"
		case inHead && !inWork && !inStaged:
			ust = "D"
		}

		if stg != "" || ust != "" {
			status.Items = append(status.Items, FileStatus{
				Path:     p,
				Staged:   stg,
				Unstaged: ust,
			})
			continue
		}

		// untracked
		if !inStaged && !inHead && inWork && opts.UntrackedMode != "no" {
			status.Untracked = append(status.Untracked, p)
		}
	}

	// ignored
	if opts.ShowIgnored {
		for _, e := range ignoredFS.Files {
			status.Ignored = append(status.Ignored, e.Path)
		}
	}

	sort.Strings(status.Untracked)
	sort.Strings(status.Ignored)

	status.IsClean = len(status.Items) == 0 &&
		len(status.Untracked) == 0 &&
		len(status.Ignored) == 0

	return &status, nil
}
