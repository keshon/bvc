package check

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/pkg/command"
	"github.com/keshon/bvc/pkg/util"
)

type WorkspaceCmd struct {
	repair bool
}

func (c *WorkspaceCmd) Name() string  { return "workspace" }
func (c *WorkspaceCmd) Brief() string { return "Verify and optionally restore working tree files" }
func (c *WorkspaceCmd) Help() string  { return "Verify and optionally restore working tree files" }
func (c *WorkspaceCmd) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.repair, "repair", false, "Restore missing or corrupted files from HEAD snapshots")
}
func (c *WorkspaceCmd) SubCommands() []command.Command { return nil }

func (c *WorkspaceCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	fmt.Println("Checking workspace files...")

	head, err := r.Head.Load()
	if err != nil {
		return fmt.Errorf("load HEAD: %w", err)
	}
	if head.Ref == "" {
		fmt.Println("HEAD is empty, nothing to verify.")
		return nil
	}

	filesToCheck := map[string]string{}

	switch head.Mode {
	case "snapshot-first":
		snap, err := r.Snaps.Load(head.Ref)
		if err != nil {
			return fmt.Errorf("load snapshot: %w", err)
		}
		for rel := range snap.Files {
			filesToCheck[rel] = head.Ref
		}

	case "stream-first":
		snapIDs, err := r.StreamSnapshots(head.Ref)
		if err != nil {
			return fmt.Errorf("load stream snapshots: %w", err)
		}

		for _, id := range snapIDs {
			snap, err := r.Snaps.Load(id)
			if err != nil {
				return err
			}
			for rel := range snap.Files {
				filesToCheck[rel] = id
			}
		}

	default:
		return fmt.Errorf("unknown HEAD mode: %s", head.Mode)
	}

	type fileEntry struct {
		rel      string
		snapshot string
	}

	var entries []fileEntry
	for rel, snapID := range filesToCheck {
		entries = append(entries, fileEntry{rel: rel, snapshot: snapID})
	}

	var mu sync.Mutex
	var missing []string

	err = util.Parallel(entries, 8, func(ctx context.Context, fe fileEntry) error {
		if _, err := os.Stat(fe.rel); os.IsNotExist(err) {
			mu.Lock()
			missing = append(missing, fe.rel)
			mu.Unlock()

			if c.repair {
				fmt.Printf("Restoring: %s\n", fe.rel)
				if err := r.CheckoutFile(fe.snapshot, fe.rel); err != nil {
					fmt.Printf("  Failed to restore: %v\n", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(missing) == 0 {
		fmt.Println("All files exist in workspace.")
	} else {
		fmt.Printf("Missing or corrupted files: %d\n", len(missing))
	}

	return nil
}
