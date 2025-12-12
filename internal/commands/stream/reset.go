package stream

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/pkg/command"
)

type ResetCmd struct {
	target string
}

func (c *ResetCmd) Name() string { return "reset" }
func (c *ResetCmd) Brief() string {
	return "Reset a stream to a specific snapshot (linear history rewrite)"
}
func (c *ResetCmd) Help() string {
	return "Reset a stream to a specific snapshot (linear history rewrite)"
}
func (c *ResetCmd) Flags(fs *flag.FlagSet)         {}
func (c *ResetCmd) SubCommands() []command.Command { return nil }
func (c *ResetCmd) Run(ctx *command.Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("usage: bvc stream reset <stream> <snapshot>")
	}
	streamName := ctx.Args[0]
	snapID := ctx.Args[1]

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	if r.Mode != "stream-first" {
		return fmt.Errorf("stream reset is allowed only in stream-first mode")
	}

	meta, err := r.Streams.Load(streamName)
	if err != nil {
		return fmt.Errorf("load stream: %w", err)
	}

	idx := -1
	for i, s := range meta.Snapshots {
		if s == snapID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("snapshot '%s' is not in stream '%s'", snapID, streamName)
	}

	newList := append([]string(nil), meta.Snapshots[:idx+1]...)
	meta.Snapshots = newList

	if err := r.Streams.Save(meta); err != nil {
		return fmt.Errorf("save stream: %w", err)
	}

	if err := r.HeadSetStream(streamName); err != nil {
		return fmt.Errorf("update HEAD: %w", err)
	}

	if err := r.CheckoutSnapshot(snapID, true); err != nil {
		return fmt.Errorf("checkout snapshot: %w", err)
	}

	fmt.Printf("Stream '%s' reset to snapshot '%s'.\n", streamName, snapID)
	fmt.Printf("New length: %d\n", len(newList))
	return nil
}
