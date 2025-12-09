package stream

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type AddCmd struct{}

func (c *AddCmd) Name() string                   { return "add" }
func (c *AddCmd) Help() string                   { return "Add snapshot to stream" }
func (c *AddCmd) Flags(fs *flag.FlagSet)         {}
func (c *AddCmd) SubCommands() []command.Command { return nil }

func (c *AddCmd) Run(ctx command.Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("usage: bvc stream add <stream> <snapshotID>")
	}
	name := ctx.Args[0]
	id := ctx.Args[1]

	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}

	_, err = repo.Snaps.Load(id)
	if err != nil {
		return fmt.Errorf("snapshot '%s' not found", id)
	}

	if err := repo.StreamAdd(name, id); err != nil {
		return err
	}

	fmt.Printf("Stream %s + %s\n", name, id)
	return nil

}
