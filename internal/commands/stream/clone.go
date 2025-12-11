package stream

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/pkg/command"
)

type CloneCmd struct{}

func (c *CloneCmd) Name() string                   { return "clone" }
func (c *CloneCmd) Brief() string                  { return "Clone a stream into a new stream" }
func (c *CloneCmd) Help() string                   { return "Clone a stream into a new stream" }
func (c *CloneCmd) Flags(fs *flag.FlagSet)         {}
func (c *CloneCmd) SubCommands() []command.Command { return nil }

func (c *CloneCmd) Run(ctx *command.Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("usage: bvc stream clone <source> <dest>")
	}
	src, dst := ctx.Args[0], ctx.Args[1]

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	if err := r.RequireMode("stream-first"); err != nil {
		return err
	}

	if _, err := r.Streams.Load(src); err != nil {
		return fmt.Errorf("source stream '%s' not found", src)
	}

	if err := r.StreamClone(src, dst); err != nil {
		return err
	}

	fmt.Printf("Stream '%s' cloned to '%s'\n", src, dst)
	return nil
}
