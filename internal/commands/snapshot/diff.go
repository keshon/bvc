package snapshot

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type DiffCmd struct{}

func (c *DiffCmd) Name() string                   { return "diff" }
func (c *DiffCmd) Help() string                   { return "Compare two snapshots" }
func (c *DiffCmd) Flags(fs *flag.FlagSet)         {}
func (c *DiffCmd) SubCommands() []command.Command { return nil }

func (c *DiffCmd) Run(ctx command.Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("usage: bvc snapshot diff <snapshotA> <snapshotB>")
	}
	aID := ctx.Args[0]
	bID := ctx.Args[1]

	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}
	if err := repo.RequireMode("snapshot-first"); err != nil {
		return err
	}

	a, err := repo.Snaps.Load(aID)
	if err != nil {
		return err
	}
	b, err := repo.Snaps.Load(bID)
	if err != nil {
		return err
	}

	// go through A
	for rel, aBlocks := range a.Files {
		bBlocks, ok := b.Files[rel]
		if !ok {
			fmt.Printf("DELETED: %s\n", rel)
			continue
		}
		if !equalStringSlices(aBlocks, bBlocks) {
			fmt.Printf("MODIFIED: %s\n", rel)
		}
	}
	// go through B for new files
	for rel := range b.Files {
		if _, ok := a.Files[rel]; !ok {
			fmt.Printf("NEW: %s\n", rel)
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
