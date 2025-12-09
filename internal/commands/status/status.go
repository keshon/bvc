package status

import (
	"bvc/command"
	"bvc/internal/repo"
	"bvc/internal/snapshot"
	"flag"
	"fmt"
	"os"
)

type StatusCmd struct{}

func (c *StatusCmd) Name() string                   { return "status" }
func (c *StatusCmd) Help() string                   { return "Show changed files since last snapshot" }
func (c *StatusCmd) Flags(fs *flag.FlagSet)         {}
func (c *StatusCmd) SubCommands() []command.Command { return nil }

func (c *StatusCmd) Run(ctx command.Context) error {

	repo, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	listIds, err := repo.Snaps.List()
	if err != nil {
		return fmt.Errorf("list snapshots: %w", err)
	}
	fmt.Printf("Found %d snapshots\n", len(listIds))
	list := make([]snapshot.Meta, len(listIds))
	for i, id := range listIds {
		meta, err := repo.Snaps.Load(id)
		if err != nil {
			return err
		}
		list[i] = *meta
	}
	if len(list) == 0 {
		fmt.Println("No snapshots yet")
		return nil
	}

	// take latest by CreatedAt
	latest := list[0]
	for _, s := range list {
		if s.CreatedAt.After(latest.CreatedAt) {
			latest = s
		}
	}

	snapFiles := latest.Files

	// collect current disk files
	currFiles := make(map[string][]string)
	err = repo.WalkFiles(func(rel string, f *os.File) error {
		hashes, err := repo.HashFileBlocks(f)
		if err != nil {
			return err
		}
		currFiles[rel] = hashes
		return nil
	})
	if err != nil {
		return err
	}

	var changes []string

	// detect new and modified
	for rel, curr := range currFiles {
		old, ok := snapFiles[rel]
		if !ok {
			changes = append(changes, fmt.Sprintf("NEW: %s", rel))
			continue
		}
		if !equalStringSlices(curr, old) {
			changes = append(changes, fmt.Sprintf("MODIFIED: %s", rel))
		}
	}

	// detect deleted
	for rel := range snapFiles {
		if _, ok := currFiles[rel]; !ok {
			changes = append(changes, fmt.Sprintf("DELETED: %s", rel))
		}
	}

	if len(changes) == 0 {
		// если изменений нет, показываем все файлы snapshot
		fmt.Println("No changes since last snapshot. Files in current snapshot:")
		for rel := range snapFiles {
			fmt.Println(rel)
		}
	} else {
		for _, line := range changes {
			fmt.Println(line)
		}
	}

	return nil
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
