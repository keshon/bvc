package stream

import (
	"bvc/command"
	"bvc/internal/repo"
	"flag"
	"fmt"
)

type ListCmd struct{}

func (c *ListCmd) Name() string                   { return "list" }
func (c *ListCmd) Help() string                   { return "List all streams" }
func (c *ListCmd) Flags(fs *flag.FlagSet)         {}
func (c *ListCmd) SubCommands() []command.Command { return nil }

func (c *ListCmd) Run(ctx command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	list, err := r.StreamList()
	if err != nil {
		return err
	}
	for _, name := range list {
		fmt.Println(name)
	}
	return nil
}
