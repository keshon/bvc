package check

import (
	"bvc/pkg/command"
	"flag"
	"fmt"
)

type CheckCmd struct{}

func (c *CheckCmd) Name() string           { return "check" }
func (c *CheckCmd) Brief() string          { return "Verify repository integrity" }
func (c *CheckCmd) Help() string           { return "Verify repository integrity" }
func (c *CheckCmd) Flags(fs *flag.FlagSet) {}
func (c *CheckCmd) SubCommands() []command.Command {
	return []command.Command{
		&BlocksCmd{},
		&WorkspaceCmd{},
	}
}
func (c *CheckCmd) Run(ctx *command.Context) error {
	fmt.Print("Available subcommands:")
	for _, sc := range c.SubCommands() {
		fmt.Print(" " + sc.Name())
	}
	fmt.Println()
	return nil
}
