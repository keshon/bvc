package prune

import (
	"bvc/internal/repo"
	"bvc/pkg/command"
	"flag"
	"fmt"
)

type PruneCmd struct {
	dryRun bool
}

func (c *PruneCmd) Name() string  { return "prune" }
func (c *PruneCmd) Brief() string { return "Remove unused blocks not referenced by any snapshot" }
func (c *PruneCmd) Help() string  { return "Remove unused blocks not referenced by any snapshot" }
func (c *PruneCmd) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.dryRun, "dry", false, "Only show what would be removed")
}

func (c *PruneCmd) SubCommands() []command.Command { return nil }

func (c *PruneCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	toDelete, err := r.Prune(c.dryRun)
	if err != nil {
		return err
	}

	if len(toDelete) == 0 {
		fmt.Println("No blocks to prune.")
		return nil
	}

	if c.dryRun {
		fmt.Println("Unused blocks (dry run):")
		for _, b := range toDelete {
			fmt.Println("  ", b)
		}
	} else {
		fmt.Printf("Pruned %d blocks successfully.\n", len(toDelete))
	}

	return nil
}
