package snapshot

import (
	"bvc/command"
	"bvc/internal/repo"
	"flag"
	"fmt"
	"time"
)

type ShowCmd struct{}

func (c *ShowCmd) Name() string                   { return "show" }
func (c *ShowCmd) Help() string                   { return "Show details of a snapshot" }
func (c *ShowCmd) Flags(fs *flag.FlagSet)         {}
func (c *ShowCmd) SubCommands() []command.Command { return nil }

func (c *ShowCmd) Run(ctx command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	id, err := r.GetHeadOrArg(ctx.Args, "snapshot:")
	if err != nil {
		return err
	}

	meta, err := r.Snaps.Load(id)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	fmt.Printf("Snapshot\n")
	fmt.Printf("  ID: %s\n", meta.ID)
	fmt.Printf("  Name: %s\n", meta.Name)
	fmt.Printf("  Created: %s\n", meta.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Description: %s\n", meta.Description)
	fmt.Printf("  Files: %d\n", len(meta.Files))
	for p, blocks := range meta.Files {
		fmt.Printf("    %s  %d blocks\n", p, len(blocks))
	}

	return nil
}
