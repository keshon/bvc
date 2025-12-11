package snapshot

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/pkg/command"
)

type MergeCmd struct{}

func (c *MergeCmd) Name() string                   { return "merge" }
func (c *MergeCmd) Brief() string                  { return "Merge two snapshots" }
func (c *MergeCmd) Help() string                   { return "Merge two snapshots" }
func (c *MergeCmd) Flags(fs *flag.FlagSet)         {}
func (c *MergeCmd) SubCommands() []command.Command { return nil }

func (c *MergeCmd) Run(ctx *command.Context) error {
	if len(ctx.Args) < 3 {
		return fmt.Errorf("usage: bvc snapshot merge <a> <b> <newSnapshotName>")
	}
	a := ctx.Args[0]
	b := ctx.Args[1]
	newName := ctx.Args[2]

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}
	if err := r.RequireMode("snapshot-first"); err != nil {
		return err
	}

	snap, err := r.MergeSnapshots(a, b, newName)
	if err != nil {
		return err
	}

	fmt.Println("Merged into snapshot:", snap.ID)
	return nil
}
