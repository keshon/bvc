package snapshot

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type MergeCmd struct{}

func (c *MergeCmd) Name() string                   { return "merge" }
func (c *MergeCmd) Help() string                   { return "Merge two snapshots" }
func (c *MergeCmd) Flags(fs *flag.FlagSet)         {}
func (c *MergeCmd) SubCommands() []command.Command { return nil }

func (c *MergeCmd) Run(ctx command.Context) error {
	if len(ctx.Args) < 3 {
		return fmt.Errorf("usage: bvc snapshot merge <a> <b> <newSnapshotName>")
	}
	a := ctx.Args[0]
	b := ctx.Args[1]
	newName := ctx.Args[2]

	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}
	if err := repo.RequireMode("snapshot-first"); err != nil {
		return err
	}

	meta, err := repo.MergeSnapshots(a, b, newName)
	if err != nil {
		return err
	}

	fmt.Println("Merged into snapshot:", meta.ID)
	return nil
}
