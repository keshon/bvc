package repo

import (
	"fmt"
	"sort"

	"github.com/keshon/bvc/internal/repo/file"
)

func (r *Repository) Add(paths []string, includeAll bool, updateOnly bool) ([]file.Entry, error) {
	// load HEAD, untracked, staged filesets
	headSet, untrackedSet, stagedSet, _, err := r.Snapshot.BuildAllRepositoryFilesets()
	if err != nil {
		return nil, fmt.Errorf("scan repository state: %w", err)
	}

	// index lookup for quick access
	indexLookup := make(map[string]file.Entry, len(stagedSet.Files))
	for _, e := range stagedSet.Files {
		indexLookup[e.Path] = e
	}

	var stagedEntries []file.Entry

	switch {
	case includeAll:
		// stage all: HEAD + untracked + already staged (deduplicated)
		allFiles := append(headSet.Files, untrackedSet.Files...)
		resultSet := make(map[string]file.Entry)

		// keep existing index entries
		for _, e := range stagedSet.Files {
			resultSet[e.Path] = e
		}

		// add all known files
		for _, f := range allFiles {
			resultSet[f.Path] = f
		}

		// convert map to slice
		stagedEntries = make([]file.Entry, 0, len(resultSet))
		for _, e := range resultSet {
			stagedEntries = append(stagedEntries, e)
		}

		// stable ordering
		sort.Slice(stagedEntries, func(i, j int) bool { return stagedEntries[i].Path < stagedEntries[j].Path })

	case updateOnly:
		// restage only the files that are both tracked and already staged
		for _, e := range headSet.Files {
			if _, wasStaged := indexLookup[e.Path]; wasStaged {
				stagedEntries = append(stagedEntries, e)
			}
		}

	default:
		// stage based on user-provided patterns
		if len(paths) == 0 {
			paths = []string{"."}
		}

		allFiles := append(headSet.Files, untrackedSet.Files...)
		matchingEntries := filterEntriesByPatterns(allFiles, paths)

		// overwrite or add matching paths to index
		for _, e := range matchingEntries {
			indexLookup[e.Path] = e
		}

		// convert map to slice
		for _, e := range indexLookup {
			stagedEntries = append(stagedEntries, e)
		}
	}

	if len(stagedEntries) == 0 {
		return nil, fmt.Errorf("no changes to stage")
	}

	if err := r.File.SaveIndexReplace(stagedEntries); err != nil {
		return nil, fmt.Errorf("update index: %w", err)
	}

	return stagedEntries, nil
}
