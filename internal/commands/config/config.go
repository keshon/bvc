package config

import (
	"bvc/command"
	"flag"
	"fmt"
)

// bvc config
type ConfigCmd struct{}

func (c *ConfigCmd) Name() string           { return "config" }
func (c *ConfigCmd) Help() string           { return "View or change configuration values" }
func (c *ConfigCmd) Flags(fs *flag.FlagSet) {}

func (c *ConfigCmd) Run(ctx command.Context) error {
	fmt.Print("Available subcommands:")
	for _, sc := range c.SubCommands() {
		fmt.Print(" " + sc.Name())
	}
	fmt.Println()
	return nil
}

func (c *ConfigCmd) SubCommands() []command.Command {
	return []command.Command{
		&SetCmd{},
		&ShowCmd{},
	}
}
