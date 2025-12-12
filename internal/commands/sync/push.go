package sync

import (
	"bvc/pkg/command"
	"flag"
	"fmt"
)

type PushCmd struct{}

func (c *PushCmd) Name() string                   { return "push" }
func (c *PushCmd) Brief() string                  { return "Push snapshots and blocks to remote storage" }
func (c *PushCmd) Help() string                   { return "Push snapshots and blocks to remote storage" }
func (c *PushCmd) Flags(fs *flag.FlagSet)         {}
func (c *PushCmd) SubCommands() []command.Command { return nil }
func (c *PushCmd) Run(ctx *command.Context) error {
	fmt.Println("Pushing to remote... (not yet implemented)")
	return nil
}
