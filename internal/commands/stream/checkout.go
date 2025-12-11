package stream

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/pkg/command"
)

type CheckoutCmd struct {
}

func (c *CheckoutCmd) Name() string                   { return "checkout" }
func (c *CheckoutCmd) Brief() string                  { return "Checkout latest snapshot from stream" }
func (c *CheckoutCmd) Help() string                   { return "Checkout latest snapshot from stream" }
func (c *CheckoutCmd) Flags(fs *flag.FlagSet)         {}
func (c *CheckoutCmd) SubCommands() []command.Command { return nil }

func (c *CheckoutCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	name, err := r.GetHeadOrArg(ctx.Args, "stream-first")
	if err != nil {
		return err
	}

	snaps, err := r.StreamSnapshots(name)
	if err != nil {
		return err
	}

	if r.Mode == "snapshot-first" {
		// sequential apply all snapshots
		for _, s := range snaps {
			if err := r.CheckoutSnapshot(s, true); err != nil { // false or true??
				return err
			}
		}
		return r.HeadSetStream(name)
	}

	// stream-first - only last snapshot, but allow empty stream
	if len(snaps) == 0 {
		fmt.Printf("Stream '%s' is empty, switching HEAD\n", name)

		if err := r.CleanupWorkdir(); err != nil {
			return err
		}

		return r.HeadSetStream(name)
	}

	latest := snaps[len(snaps)-1]
	if err := r.CheckoutSnapshot(latest, true); err != nil {
		return err
	}
	return r.HeadSetStream(name)
}
