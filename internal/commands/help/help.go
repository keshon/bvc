package help

import (
	"flag"
	"fmt"

	"bvc/pkg/command"
)

type HelpCmd struct {
	Reg *command.Registry
}

func NewHelpCmd(reg *command.Registry) *HelpCmd {
	return &HelpCmd{Reg: reg}
}

func (c *HelpCmd) Name() string           { return "help" }
func (c *HelpCmd) Help() string           { return "Show help for all commands" }
func (c *HelpCmd) Flags(fs *flag.FlagSet) {}

func (c *HelpCmd) SubCommands() []command.Command { return nil }

func (c *HelpCmd) Run(ctx command.Context) error {
	fmt.Println("Available commands:")

	all := c.Reg.GetAll()

	for _, cmd := range all {
		name := cmd.Name()
		fmt.Printf("  %s - %s\n", name, cmd.Help())

		for _, sc := range cmd.SubCommands() {
			fmt.Printf("    %s %s - %s\n", name, sc.Name(), sc.Help())
		}
	}

	return nil
}
