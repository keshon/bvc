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

	// Получаем все файлы
	trackedFS, untrackedFS, stagedFS, ignoredFS, err := r.Snapshot.BuildAllRepositoryFilesets()
	if err != nil {
		return nil, fmt.Errorf("scan working tree: %w", err)
	}

	status := RepoStatus{Branch: branch.Name}

	// Мапы для быстрого поиска
	headMap := make(map[string]file.Entry)
	for _, e := range trackedFS.Files {
		headMap[e.Path] = e
	}

	stagedMap := make(map[string]file.Entry)
	for _, e := range stagedFS.Files {
		stagedMap[e.Path] = e
	}

	// Получаем все пути: из HEAD, индекса и рабочей директории
	allPathsMap := make(map[string]struct{})
	for _, e := range trackedFS.Files {
		allPathsMap[e.Path] = struct{}{}
	}
	for _, e := range stagedFS.Files {
		allPathsMap[e.Path] = struct{}{}
	}
	for _, e := range untrackedFS.Files {
		allPathsMap[e.Path] = struct{}{}
	}

	var allPaths []string
	for p := range allPathsMap {
		allPaths = append(allPaths, p)
	}
	sort.Strings(allPaths)

	for _, path := range allPaths {
		h, inHead := headMap[path]
		s, inStaged := stagedMap[path]
		fileExists := containsFile(trackedFS.Files, path) || r.FS.Exists(path)

		var stg, ust string

		// Staged
		switch {
		case inStaged && !inHead:
			stg = "A" // новый файл
		case inStaged && inHead && !h.Equal(&s):
			stg = "M" // модифицирован
		case inHead && !inStaged && !fileExists:
			stg = "D" // удалён
		}

		// Unstaged
		if inHead && fileExists {
			entryInWD := getEntry(trackedFS.Files, path)
			if !inStaged && entryInWD != nil && !h.Equal(entryInWD) {
				ust = "M"
			}
		}

		if stg != "" || ust != "" {
			status.Items = append(status.Items, FileStatus{
				Path:     path,
				Staged:   stg,
				Unstaged: ust,
			})
		}
	}

	// Untracked
	if opts.UntrackedMode != "no" {
		for _, e := range untrackedFS.Files {
			status.Untracked = append(status.Untracked, e.Path)
		}
	}

	// Ignored
	if opts.ShowIgnored {
		for _, e := range ignoredFS.Files {
			status.Ignored = append(status.Ignored, e.Path)
		}
	}

	sort.Strings(status.Untracked)
	sort.Strings(status.Ignored)
	status.IsClean = len(status.Items) == 0 && len(status.Untracked) == 0 && len(status.Ignored) == 0

	return &status, nil
}

// helper: найти Entry в слайсе по пути
func getEntry(files []file.Entry, path string) *file.Entry {
	path = filepath.Clean(path)
	for i := range files {
		if filepath.Clean(files[i].Path) == path {
			return &files[i]
		}
	}
	return nil
}

// helper: проверить, содержится ли Entry в слайсе по пути
func containsFile(files []file.Entry, path string) bool {
	path = filepath.Clean(path)
	for i := range files {
		if filepath.Clean(files[i].Path) == path {
			return true
		}
	}
	return false
}
