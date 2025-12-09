package snapshot

import (
	"bvc/command"
	"bvc/internal/repo"
	"flag"
	"fmt"
	"time"
)

type CreateCmd struct {
	name        string
	description string
}

func (c *CreateCmd) Name() string { return "create" }
func (c *CreateCmd) Help() string { return "Create a new snapshot of the project" }

func (c *CreateCmd) Flags(fs *flag.FlagSet) {
	fs.StringVar(&c.name, "name", "", "Snapshot name (optional)")
	fs.StringVar(&c.description, "desc", "", "Description of the snapshot")
}

func (c *CreateCmd) SubCommands() []command.Command { return nil }

func (c *CreateCmd) Run(ctx command.Context) error {
	if c.name == "" {
		c.name = "snap-" + time.Now().Format("20060102-150405")
	}

	repo, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	fmt.Printf("Creating snapshot '%s' with description '%s'\n", c.name, c.description)
	meta, err := repo.CreateSnapshot(c.name, c.description)
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}

	if repo.Mode == "snapshot-first" {
		if err := repo.HeadSetSnapshot(meta.ID); err != nil {
			return err
		}
	}

	fmt.Printf("Snapshot created: id=%s created_at=%s files=%d\n",
		meta.ID, meta.CreatedAt.Format(time.RFC3339), len(meta.Files))

	for p, hashes := range meta.Files {
		fmt.Printf("  %s (%d blocks)\n", p, len(hashes))
	}
	return nil
}
