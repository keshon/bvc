package snapshot

import (
	"bvc/command"
	"flag"
	"fmt"
)

// snapshot root command
type SnapshotCmd struct{}

func (c *SnapshotCmd) Name() string           { return "snapshot" }
func (c *SnapshotCmd) Help() string           { return "Manage project snapshots" }
func (c *SnapshotCmd) Flags(fs *flag.FlagSet) {}
func (c *SnapshotCmd) Run(ctx command.Context) error {
	fmt.Print("Available subcommands:")
	for _, sc := range c.SubCommands() {
		fmt.Print(" " + sc.Name())
	}
	fmt.Println()
	return nil
}
func (c *SnapshotCmd) SubCommands() []command.Command {
	return []command.Command{
		&CheckoutCmd{},
		&CreateCmd{},
		&DiffCmd{},
		&ListCmd{},
		&MergeCmd{},
		&ShowCmd{},
	}
}
