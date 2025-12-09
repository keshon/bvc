package snapshot

import (
	"bvc/command"
	"bvc/internal/repo"
	"flag"
	"fmt"
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

	fmt.Printf("Snapshot %s\nName: %s\nCreated: %s\nDescription: %s\nFiles: %d\n",
		meta.ID, meta.Name, meta.CreatedAt, meta.Description, len(meta.Files))
	for p, blocks := range meta.Files {
		fmt.Printf("  %s (%d blocks)\n", p, len(blocks))
	}
	return nil
}
