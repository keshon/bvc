package stream

import (
	"flag"
	"fmt"

	"bvc/pkg/command"
)

type StreamCmd struct{}

func (c *StreamCmd) Name() string           { return "stream" }
func (c *StreamCmd) Brief() string          { return "Manage snapshot streams (catalogs)" }
func (c *StreamCmd) Help() string           { return "Manage snapshot streams (catalogs)" }
func (c *StreamCmd) Flags(fs *flag.FlagSet) {}
func (c *StreamCmd) SubCommands() []command.Command {
	return []command.Command{
		&CreateCmd{},
		&AddCmd{},
		&ListCmd{},
		&ShowCmd{},
		&CheckoutCmd{},
		&RemoveCmd{},
	}
}
func (c *StreamCmd) Run(ctx *command.Context) error {
	fmt.Print("Available subcommands:")
	for _, sc := range c.SubCommands() {
		fmt.Print(" " + sc.Name())
	}
	fmt.Println()
	return nil
}
