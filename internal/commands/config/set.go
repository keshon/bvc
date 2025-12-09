package config

import (
	"bvc/command"
	"flag"
	"fmt"
)

type SetCmd struct{}

func (c *SetCmd) Name() string                   { return "set" }
func (c *SetCmd) Help() string                   { return "Set a configuration value" }
func (c *SetCmd) Flags(fs *flag.FlagSet)         {}
func (c *SetCmd) SubCommands() []command.Command { return nil }

func (c *SetCmd) Run(ctx command.Context) error {
	if len(ctx.Args) < 2 {
		return fmt.Errorf("usage: bvc config set <key> <value>")
	}

	key := ctx.Args[0]
	value := ctx.Args[1]
	fmt.Printf("Setting config %s = %s\n", key, value)

	// TODO: ConfigManager.Set(key, value)

	return nil
}
