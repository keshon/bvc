package stream

import (
	"bvc/command"
	"bvc/internal/repo"
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

	repo, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	if err := repo.StreamCreate(name); err != nil {
		return err
	}

	if repo.Mode == "stream-first" {
		if err := repo.HeadSetStream(name); err != nil {
			return err
		}
	}

	fmt.Printf("Stream '%s' created\n", name)
	return nil
}
