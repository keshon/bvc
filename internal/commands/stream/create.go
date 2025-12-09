package stream

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type CreateCmd struct{}

func (c *CreateCmd) Name() string                   { return "create" }
func (c *CreateCmd) Help() string                   { return "Create a new stream" }
func (c *CreateCmd) Flags(fs *flag.FlagSet)         {}
func (c *CreateCmd) SubCommands() []command.Command { return nil }

func (c *CreateCmd) Run(ctx command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: bvc stream create <name>")
	}
	name := ctx.Args[0]

	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}

	if err := repo.StreamCreate(name); err != nil {
		return err
	}

	fmt.Printf("Stream '%s' created\n", name)
	return nil

}
