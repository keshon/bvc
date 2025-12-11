package sync

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/pkg/command"
)

type SyncCmd struct{}

func (c *SyncCmd) Name() string           { return "sync" }
func (c *SyncCmd) Brief() string          { return "Synchronize snapshots and blocks with remote" }
func (c *SyncCmd) Help() string           { return "Synchronize snapshots and blocks with remote" }
func (c *SyncCmd) Flags(fs *flag.FlagSet) {}
func (c *SyncCmd) Run(ctx *command.Context) error {
	fmt.Print("Available subcommands:")
	for _, sc := range c.SubCommands() {
		fmt.Print(" " + sc.Name())
	}
	fmt.Println()
	return nil
}
func (c *SyncCmd) SubCommands() []command.Command {
	return []command.Command{
		&PullCmd{},
		&PushCmd{},
	}
}
