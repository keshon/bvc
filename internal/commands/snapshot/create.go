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

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	snap, err := r.CreateSnapshot(c.name, c.description)
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}

	switch r.Mode {
	case "snapshot-first":
		if err := r.HeadSetSnapshot(snap.ID); err != nil {
			return err
		}
	case "stream-first":
		head, err := r.Head.Load()
		if err != nil {
			return fmt.Errorf("load HEAD: %w", err)
		}
		if head.Mode != "stream-first" || head.Ref == "" {
			return fmt.Errorf("HEAD is not pointing to a stream")
		}
		if err := r.StreamAdd(head.Ref, snap.ID); err != nil {
			return fmt.Errorf("add snapshot to current stream: %w", err)
		}
	default:
		return fmt.Errorf("unsupported mode: %s", r.Mode)
	}

	fmt.Printf("Snapshot created\n")
	fmt.Printf("  ID: %s\n", snap.ID)
	fmt.Printf("  Name: %s\n", snap.Name)
	fmt.Printf("  Created: %s\n", snap.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Description: %s\n", snap.Description)
	fmt.Printf("  Files: %d\n", len(snap.Files))
	for p, hashes := range snap.Files {
		fmt.Printf("    %s  %d blocks\n", p, len(hashes))
	}

	return nil
}
