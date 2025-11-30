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

	// build all Filesets: HEAD, index, untracked, ignored
	headSet, untrackedSet, indexSet, ignoredSet, err := r.Snapshot.BuildAllRepositoryFilesets()
	if err != nil {
		return nil, fmt.Errorf("scan working tree: %w", err)
	}

	status := RepoStatus{Branch: branch.Name}

	// lookup tables for quick comparisons
	headByPath := make(map[string]file.Entry)
	for _, e := range headSet.Files {
		headByPath[e.Path] = e
	}

	indexByPath := make(map[string]file.Entry)
	for _, e := range indexSet.Files {
		indexByPath[e.Path] = e
	}

	// collect all unique paths appearing in HEAD, index, or untracked
	uniquePaths := make(map[string]struct{})
	for _, e := range headSet.Files {
		uniquePaths[e.Path] = struct{}{}
	}
	for _, e := range indexSet.Files {
		uniquePaths[e.Path] = struct{}{}
	}
	for _, e := range untrackedSet.Files {
		uniquePaths[e.Path] = struct{}{}
	}

	var allPaths []string
	for p := range uniquePaths {
		allPaths = append(allPaths, p)
	}
	sort.Strings(allPaths)

	for _, path := range allPaths {
		headEntry, existsInHead := headByPath[path]
		indexEntry, existsInIndex := indexByPath[path]
		existsInWorkingTree := containsPath(headSet.Files, path) || r.FS.Exists(path)

		var stagedStatus, unstagedStatus string

		// staged status
		switch {
		case existsInIndex && !existsInHead:
			stagedStatus = "A" // new file
		case existsInIndex && existsInHead && !headEntry.Equal(&indexEntry):
			stagedStatus = "M" // modified
		case existsInHead && !existsInIndex && !existsInWorkingTree:
			stagedStatus = "D" // deleted
		}

		// unstaged status
		if existsInHead && existsInWorkingTree {
			entryInWD := findEntry(headSet.Files, path)
			if !existsInIndex && entryInWD != nil && !headEntry.Equal(entryInWD) {
				unstagedStatus = "M"
			}
		}

		if stagedStatus != "" || unstagedStatus != "" {
			status.Items = append(status.Items, FileStatus{
				Path:     path,
				Staged:   stagedStatus,
				Unstaged: unstagedStatus,
			})
		}
	}

	// untracked files
	if opts.UntrackedMode != "no" {
		for _, e := range untrackedSet.Files {
			status.Untracked = append(status.Untracked, e.Path)
		}
	}

	// ignored files
	if opts.ShowIgnored {
		for _, e := range ignoredSet.Files {
			status.Ignored = append(status.Ignored, e.Path)
		}
	}

	sort.Strings(status.Untracked)
	sort.Strings(status.Ignored)

	status.IsClean =
		len(status.Items) == 0 &&
			len(status.Untracked) == 0 &&
			len(status.Ignored) == 0

	return &status, nil
}

// findEntry returns the Entry with the given path, or nil if not found.
func findEntry(files []file.Entry, path string) *file.Entry {
	clean := filepath.Clean(path)
	for i := range files {
		if filepath.Clean(files[i].Path) == clean {
			return &files[i]
		}
	}
	return nil
}

// containsPath returns true if any Entry has the given path.
func containsPath(files []file.Entry, path string) bool {
	clean := filepath.Clean(path)
	for i := range files {
		if filepath.Clean(files[i].Path) == clean {
			return true
		}
	}
	return false
}
