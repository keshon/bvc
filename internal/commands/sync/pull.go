package sync

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/pkg/command"
)

type PullCmd struct{}

func (c *PullCmd) Name() string                   { return "pull" }
func (c *PullCmd) Brief() string                  { return "Pull snapshots and blocks from remote storage" }
func (c *PullCmd) Help() string                   { return "Pull snapshots and blocks from remote storage" }
func (c *PullCmd) Flags(fs *flag.FlagSet)         {}
func (c *PullCmd) SubCommands() []command.Command { return nil }
func (c *PullCmd) Run(ctx *command.Context) error {
	fmt.Println("Pulling from remote... (not yet implemented)")
	return nil
}
