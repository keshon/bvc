package snapshot

import (
	"flag"
	"fmt"

	"bvc/internal/repo"
	"bvc/pkg/command"
)

type CheckoutCmd struct{}

func (c *CheckoutCmd) Name() string { return "checkout" }
func (c *CheckoutCmd) Help() string { return "Restore project from snapshot" }

func (c *CheckoutCmd) Flags(fs *flag.FlagSet) {}

func (c *CheckoutCmd) SubCommands() []command.Command { return nil }

func (c *CheckoutCmd) Run(ctx command.Context) error {

	if len(ctx.Args) < 1 {
		return fmt.Errorf("snapshot ID is required")
	}
	id := ctx.Args[0]

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	fmt.Printf("Checking out snapshot '%s'\n", id)
	if err := r.CheckoutSnapshot(id, true); err != nil {
		return fmt.Errorf("checkout: %w", err)
	}

	if r.Mode == "snapshot-first" {
		if err := r.HeadSetSnapshot(id); err != nil {
			return err
		}
	}

	fmt.Println("Checkout complete.")
	return nil
}
