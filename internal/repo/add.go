package repo

import (
	"fmt"

	"github.com/keshon/bvc/internal/repo/file"
)

func (r *Repository) Add(paths []string, includeAll bool, updateOnly bool) ([]file.Entry, error) {
	// Scan repository state
	trackedFS, stagedFS, _, err := r.Snapshot.BuildAllRepositoryFilesets()
	if err != nil {
		return nil, fmt.Errorf("scan repository state: %w", err)
	}

	var entries []file.Entry

	switch {
	case includeAll:
		// Stage all tracked (new + modified + deletions)
		entries = trackedFS.Files

	case updateOnly:
		// Only updates for previously-staged paths
		stagedMap := make(map[string]file.Entry, len(stagedFS.Files))
		for _, e := range stagedFS.Files {
			stagedMap[e.Path] = e
		}
		for _, e := range trackedFS.Files {
			if _, ok := stagedMap[e.Path]; ok {
				entries = append(entries, e)
			}
		}

	default:
		// Resolve file patterns / arguments
		if len(paths) == 0 {
			paths = []string{"."}
		}
		entries = filterEntriesByPatterns(trackedFS.Files, paths)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no changes to stage")
	}

	if err := r.File.SaveIndexMerge(entries); err != nil {
		return nil, fmt.Errorf("update index: %w", err)
	}

	return entries, nil
}
