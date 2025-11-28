package repo

import (
	"fmt"
	"sort"

	"github.com/keshon/bvc/internal/repo/file"
)

func (r *Repository) Add(paths []string, includeAll bool, updateOnly bool) ([]file.Entry, error) {
	// Получаем все файловые сеты
	trackedFS, untrackedFS, stagedFS, _, err := r.Snapshot.BuildAllRepositoryFilesets()
	if err != nil {
		return nil, fmt.Errorf("scan repository state: %w", err)
	}

	stagedMap := make(map[string]file.Entry, len(stagedFS.Files))
	for _, e := range stagedFS.Files {
		stagedMap[e.Path] = e
	}

	var entries []file.Entry

	switch {
	case includeAll:
		// Stage всё: tracked + untracked + уже staged, без дубликатов
		allFiles := append(trackedFS.Files, untrackedFS.Files...)
		fileSet := make(map[string]file.Entry)

		// Сначала добавляем уже staged
		for _, e := range stagedFS.Files {
			fileSet[e.Path] = e
		}

		// Потом новые из tracked + untracked
		for _, f := range allFiles {
			fileSet[f.Path] = f
		}

		// Собираем список
		entries = make([]file.Entry, 0, len(fileSet))
		for _, e := range fileSet {
			entries = append(entries, e)
		}

		// Сортировка для стабильности индекса
		sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })

	case updateOnly:
		// Stage tracked files that are already staged
		for _, e := range trackedFS.Files {
			if _, ok := stagedMap[e.Path]; ok {
				entries = append(entries, e)
			}
		}

	default:
		// Stage по аргументам
		if len(paths) == 0 {
			paths = []string{"."}
		}

		allFiles := append(trackedFS.Files, untrackedFS.Files...)
		selected := filterEntriesByPatterns(allFiles, paths)
		for _, e := range selected {
			stagedMap[e.Path] = e
		}

		for _, e := range stagedMap {
			entries = append(entries, e)
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no changes to stage")
	}

	if err := r.File.SaveIndexReplace(entries); err != nil {
		return nil, fmt.Errorf("update index: %w", err)
	}

	return entries, nil
}
