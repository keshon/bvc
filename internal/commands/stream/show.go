package stream

import (
	"flag"
	"fmt"

	"bvc/internal/repo"
	"bvc/pkg/command"
)

type ShowCmd struct{}

func (c *ShowCmd) Name() string                   { return "show" }
func (c *ShowCmd) Brief() string                  { return "Show snapshots in stream" }
func (c *ShowCmd) Help() string                   { return "Show snapshots in stream" }
func (c *ShowCmd) Flags(fs *flag.FlagSet)         {}
func (c *ShowCmd) SubCommands() []command.Command { return nil }

func (c *ShowCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	name, err := r.GetHeadOrArg(ctx.Args, "stream-first")
	if err != nil {
		return err
	}

	ids, err := r.StreamSnapshots(name)
	if err != nil {
		return err
	}

	fmt.Printf("Stream '%s':\n", name)
	for _, id := range ids {
		fmt.Println("  " + id)
	}
	return nil

}
