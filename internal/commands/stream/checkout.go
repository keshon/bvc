package stream

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type CheckoutCmd struct {
}

func (c *CheckoutCmd) Name() string                   { return "checkout" }
func (c *CheckoutCmd) Help() string                   { return "Checkout latest snapshot from stream" }
func (c *CheckoutCmd) Flags(fs *flag.FlagSet)         {}
func (c *CheckoutCmd) SubCommands() []command.Command { return nil }

func (c *CheckoutCmd) Run(ctx command.Context) error {
	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}

	name, err := repo.GetHeadOrArg(ctx.Args, "stream:")

	snaps, err := repo.StreamSnapshots(name)
	if err != nil {
		return err
	}

	if repo.Mode == "snapshot-first" {
		// sequential apply all snapshots
		for _, s := range snaps {
			if err := repo.CheckoutSnapshot(s, true); err != nil { // false or true??
				return err
			}
		}
		return repo.HeadSetStream("stream:" + name)
	}

	// stream-first => only last snapshot, but allow empty stream
	if len(snaps) == 0 {
		fmt.Printf("Stream '%s' is empty, switching HEAD\n", name)

		// clean workdir (git style)
		if err := repo.CleanupWorkdir(); err != nil {
			return err
		}

		return repo.HeadSetStream("stream:" + name)
	}

	latest := snaps[len(snaps)-1]
	if err := repo.CheckoutSnapshot(latest, true); err != nil {
		return err
	}
	return repo.HeadSetStream("stream:" + name)
}
