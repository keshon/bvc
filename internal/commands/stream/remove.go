package stream

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type RemoveCmd struct{}

func (c *RemoveCmd) Name() string                   { return "remove" }
func (c *RemoveCmd) Help() string                   { return "Remove a stream" }
func (c *RemoveCmd) Flags(fs *flag.FlagSet)         {}
func (c *RemoveCmd) SubCommands() []command.Command { return nil }

func (c *RemoveCmd) Run(ctx command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: bvc stream remove <name>")
	}
	name := ctx.Args[0]

	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}
	if err := repo.RequireMode("stream-first"); err != nil {
		return err
	}

	if err := repo.StreamRemove(name); err != nil {
		return err
	}

	fmt.Printf("Stream '%s' removed\n", name)
	return nil

}
