package snapshot

import (
	"flag"
	"fmt"
	"time"

	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/pkg/command"
)

type ShowCmd struct{}

func (c *ShowCmd) Name() string                   { return "show" }
func (c *ShowCmd) Brief() string                  { return "Show details of a snapshot" }
func (c *ShowCmd) Help() string                   { return "Show details of a snapshot" }
func (c *ShowCmd) Flags(fs *flag.FlagSet)         {}
func (c *ShowCmd) SubCommands() []command.Command { return nil }

func (c *ShowCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	id, err := r.GetHeadOrArg(ctx.Args, "snapshot-first")
	if err != nil {
		return err
	}

	snap, err := r.Snaps.Load(id)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	fmt.Printf("Snapshot\n")
	fmt.Printf("  ID: %s\n", snap.ID)
	fmt.Printf("  Name: %s\n", snap.Name)
	fmt.Printf("  Created: %s\n", snap.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Description: %s\n", snap.Description)
	fmt.Printf("  Files: %d\n", len(snap.Files))
	for p, blocks := range snap.Files {
		fmt.Printf("    %s  %d blocks\n", p, len(blocks))
	}

	return nil
}
