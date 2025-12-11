package stream

import (
	"flag"
	"fmt"

	"bvc/internal/repo"
	"bvc/pkg/command"
)

type AddCmd struct{}

func (c *AddCmd) Name() string                   { return "add" }
func (c *AddCmd) Brief() string                  { return "Add snapshot to stream" }
func (c *AddCmd) Help() string                   { return "Add snapshot to stream" }
func (c *AddCmd) Flags(fs *flag.FlagSet)         {}
func (c *AddCmd) SubCommands() []command.Command { return nil }

func (c *AddCmd) Run(ctx *command.Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("usage: bvc stream add <stream> <snapshotID>")
	}
	name := ctx.Args[0]
	id := ctx.Args[1]

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	_, err = r.Snaps.Load(id)
	if err != nil {
		return fmt.Errorf("snapshot '%s' not found", id)
	}

	if err := r.StreamAdd(name, id); err != nil {
		return err
	}

	fmt.Printf("Stream %s + %s\n", name, id)
	return nil

}
