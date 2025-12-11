package snapshot

import (
	"flag"
	"fmt"

	"bvc/internal/repo"
	"bvc/pkg/command"
	"bvc/pkg/util"
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

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}
	if err := r.RequireMode("snapshot-first"); err != nil {
		return err
	}

	a, err := r.Snaps.Load(aID)
	if err != nil {
		return err
	}
	b, err := r.Snaps.Load(bID)
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
		if !util.EqualStringSlices(aBlocks, bBlocks) {
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
