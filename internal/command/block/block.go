package block

import (
	"app/internal/command"
	"flag"
	"fmt"
)

// Base command for "block"
type BlockCommand struct{}

func (c *BlockCommand) Name() string      { return "block" }
func (c *BlockCommand) Brief() string     { return "Block-related commands" }
func (c *BlockCommand) Usage() string     { return "block <subcommand> [options]" }
func (c *BlockCommand) Help() string      { return "Manage repository blocks and analysis" }
func (c *BlockCommand) Aliases() []string { return []string{"bl"} }

// Subcommands now include analyze, overview, scan, and repair
func (c *BlockCommand) Subcommands() []command.Command {
	return []command.Command{
		&ReuseCommand{},
		&ListCommand{},
		&ScanCommand{},
		&RepairCommand{},
	}
}

func (c *BlockCommand) Flags(fs *flag.FlagSet) {}

// Run prints usage if no subcommand is provided
func (c *BlockCommand) Run(ctx *command.Context) error {
	fmt.Println("Usage: block <subcommand> [options]\n")
	fmt.Println("Available subcommands:\n")

	subcmds := c.Subcommands()
	longest := 0
	for _, sc := range subcmds {
		if l := len(sc.Name()); l > longest {
			longest = l
		}
	}

	for _, sc := range subcmds {
		padding := ""
		if longest > len(sc.Name()) {
			padding = "  " + spaces(longest-len(sc.Name()))
		} else {
			padding = "  "
		}
		fmt.Printf("  \033[1m%s\033[0m%s%s\n", sc.Name(), padding, sc.Brief())
	}
	fmt.Println("\nType 'block <subcommand> --help' for detailed usage of a subcommand.")
	return nil
}

func spaces(n int) string {
	return fmt.Sprintf("%*s", n, "")
}
