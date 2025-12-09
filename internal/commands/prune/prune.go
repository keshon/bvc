package prune

import (
	"bvc/command"
	"flag"
	"fmt"
)

type PruneCmd struct {
	dryRun bool
}

func (c *PruneCmd) Name() string { return "prune" }
func (c *PruneCmd) Help() string { return "Remove unused blocks not referenced by any snapshot" }

func (c *PruneCmd) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.dryRun, "dry", false, "Only show what would be removed")
}

func (c *PruneCmd) SubCommands() []command.Command { return nil }

func (c *PruneCmd) Run(ctx command.Context) error {
	if c.dryRun {
		fmt.Println("Dry run: showing unused blocks...")
	} else {
		fmt.Println("Pruning unused blocks...")
	}

	// TODO: реализация BlockStore.Prune

	return nil
}
