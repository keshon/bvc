package stream

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type ListCmd struct{}

func (c *ListCmd) Name() string                   { return "list" }
func (c *ListCmd) Help() string                   { return "List all streams" }
func (c *ListCmd) Flags(fs *flag.FlagSet)         {}
func (c *ListCmd) SubCommands() []command.Command { return nil }

func (c *ListCmd) Run(ctx command.Context) error {
	repo, err := util.OpenRepo(".")
	if err != nil {
		return err
	}

	list, err := repo.StreamList()
	if err != nil {
		return err
	}
	for _, name := range list {
		fmt.Println(name)
	}
	return nil
}
