package help

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/middleware"
)

type Command struct{}

func (c *Command) Name() string      { return "help" }
func (c *Command) Short() string     { return "H" }
func (c *Command) Aliases() []string { return []string{"h", "?"} }
func (c *Command) Usage() string     { return "help [command]" }
func (c *Command) Brief() string     { return "Show help for commands" }
func (c *Command) Help() string {
	return `Display help information for commands.

Usage:
  help          List all commands.
  help <name>   Show detailed help for a specific command.`
}

func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet)         {}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) > 0 {
		return runCommandHelp(strings.ToLower(ctx.Args[0]))
	}
	return runListAllCommands()
}

// runCommandHelp shows detailed help for a specific command
func runCommandHelp(name string) error {
	cmd, ok := command.GetCommand(name)
	if !ok {
		fmt.Printf("Unknown command: %s\n", name)
		return nil
	}

	if usage := cmd.Usage(); usage != "" {
		fmt.Printf("\033[90mUsage:\033[0m %s\n\n", usage)
	}
	fmt.Printf("%s\n\n", cmd.Help())

	if aliasesCmd, ok := cmd.(interface{ Aliases() []string }); ok {
		if aliases := aliasesCmd.Aliases(); len(aliases) > 0 {
			fmt.Printf("Aliases: %s\n", strings.Join(aliases, ", "))
		}
	}

	return nil
}

// runListAllCommands lists all commands in a Git-style layout
func runListAllCommands() error {
	commands := command.AllCommands()
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	fmt.Print("Available commands:\n\n")
	longest := 0
	for _, cmd := range commands {
		if l := len(cmd.Name()); l > longest {
			longest = l
		}
	}

	for _, cmd := range commands {
		name := cmd.Name()
		desc := cmd.Brief()
		if desc == "" {
			desc = "-"
		}

		padding := strings.Repeat(" ", longest-len(name)+2)
		fmt.Printf("  \033[1m%s\033[0m%s%s\n", name, padding, desc)
	}

	fmt.Println("\nType 'help <command>' to see detailed information about a specific command.")
	return nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
