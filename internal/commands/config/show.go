package config

import (
	"bvc/command"
	"flag"
	"fmt"
)

type ShowCmd struct{}

func (c *ShowCmd) Name() string                   { return "show" }
func (c *ShowCmd) Help() string                   { return "Display current configuration" }
func (c *ShowCmd) Flags(fs *flag.FlagSet)         {}
func (c *ShowCmd) SubCommands() []command.Command { return nil }

func (c *ShowCmd) Run(ctx command.Context) error {
	fmt.Println("Showing configuration...")

	// TODO: ConfigManager.Show()

	return nil
}
