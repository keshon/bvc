package stream

import (
	"bvc/command"
	"bvc/internal/repo"
	"flag"
	"fmt"
)

type CreateCmd struct{}

func (c *CreateCmd) Name() string                   { return "create" }
func (c *CreateCmd) Help() string                   { return "Create a new stream" }
func (c *CreateCmd) Flags(fs *flag.FlagSet)         {}
func (c *CreateCmd) SubCommands() []command.Command { return nil }
func (c *CreateCmd) Run(ctx command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("usage: bvc stream create <name>")
	}
	name := ctx.Args[0]

	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	switch r.Mode {

	case "stream-first":
		head, err := r.Head.Load()
		if err != nil {
			return err
		}

		// HEAD empty: create empty stream
		if head.Ref == "" {
			if err := r.StreamCreate(name); err != nil {
				return err
			}
			if err := r.HeadSetStream(name); err != nil {
				return err
			}
			fmt.Printf("Stream '%s' created\n", name)
			return nil
		}

		// HEAD must be stream-first
		if head.Mode != "stream-first" {
			return fmt.Errorf("HEAD is not pointing to a stream")
		}

		// Clone from HEAD stream
		if err := r.StreamClone(head.Ref, name); err != nil {
			return err
		}
		if err := r.HeadSetStream(name); err != nil {
			return err
		}
		fmt.Printf("Stream '%s' created (cloned from '%s')\n", name, head.Ref)
		return nil

	case "snapshot-first":
		if err := r.StreamCreate(name); err != nil {
			return err
		}
		fmt.Printf("Stream '%s' created\n", name)
		return nil

	default:
		return fmt.Errorf("invalid repository mode: %s", r.Mode)
	}
}
