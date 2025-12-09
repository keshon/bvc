package sync

import (
	"bvc/command"
	"flag"
	"fmt"
)

type PullCmd struct{}

func (c *PullCmd) Name() string                   { return "pull" }
func (c *PullCmd) Help() string                   { return "Pull snapshots and blocks from remote storage" }
func (c *PullCmd) Flags(fs *flag.FlagSet)         {}
func (c *PullCmd) SubCommands() []command.Command { return nil }
func (c *PullCmd) Run(ctx command.Context) error {
	fmt.Println("Pulling from remote...")
	return nil
}
