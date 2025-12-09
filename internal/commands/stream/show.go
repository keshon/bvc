package stream

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type ShowCmd struct{}

func (c *ShowCmd) Name() string                   { return "show" }
func (c *ShowCmd) Help() string                   { return "Show snapshots in stream" }
func (c *ShowCmd) Flags(fs *flag.FlagSet)         {}
func (c *ShowCmd) SubCommands() []command.Command { return nil }

func (c *ShowCmd) Run(ctx command.Context) error {
	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}

	name, err := repo.GetHeadOrArg(ctx.Args, "stream:")
	if err != nil {
		return err
	}

	ids, err := repo.StreamSnapshots(name)
	if err != nil {
		return err
	}

	fmt.Printf("Stream '%s':\n", name)
	for _, id := range ids {
		fmt.Println("  " + id)
	}
	return nil

}
