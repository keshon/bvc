package help

import (
	"fmt"
	"sort"
	"strings"

	"app/internal/cli"
)

// Command shows help information for commands
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "help" }

// Usage string
func (c *Command) Usage() string { return "help [command]" }

// Short description
func (c *Command) Description() string { return "Show help for commands" }

// Detailed description
func (c *Command) DetailedDescription() string {
	return "Display detailed help information for a specific command, or list all commands if none is provided."
}

// Aliases
func (c *Command) Aliases() []string { return []string{"h", "?"} }

// Shortcut
func (c *Command) Short() string { return "H" }

// Run executes the help command
func (c *Command) Run(ctx *cli.Context) error {
	if len(ctx.Args) > 0 {
		return commandHelp(strings.ToLower(ctx.Args[0]))
	}
	return listAllCommands()
}

// commandHelp shows detailed help for a single command
func commandHelp(name string) error {
	cmd, ok := cli.GetCommand(name)
	if !ok {
		fmt.Printf("Unknown command: %s\n", name)
		return nil
	}

	if usage := cmd.Usage(); usage != "" {
		fmt.Printf("\033[90mUsage:\033[0m %s\n\n", usage)
	}
	fmt.Printf("%s\n\n", cmd.DetailedDescription())

	// Show aliases if available
	if aliasesCmd, ok := cmd.(interface{ Aliases() []string }); ok {
		aliases := aliasesCmd.Aliases()
		if len(aliases) > 0 {
			fmt.Printf("Aliases: %s\n", strings.Join(aliases, ", "))
		}
	}

	return nil
}

// listAllCommands lists all registered commands
func listAllCommands() error {
	commands := cli.AllCommands()
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	fmt.Println("Available commands")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))
	fmt.Printf("\033[90m%-10s %-54s %-32s\033[0m\n", "Command", "Description", "Usage")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))

	for _, cmd := range commands {
		usage := cmd.Usage()
		if usage == "" {
			usage = "-"
		}
		desc := cmd.Description()
		if desc == "" {
			desc = "-"
		}

		fmt.Printf("%-10s \033[90m%-54s\033[0m %-32s\n", cmd.Name(), desc, usage)
	}

	fmt.Println("\nUse 'help <command>' to see detailed usage of a specific command.")
	return nil
}

// Register command
func init() {
	cli.RegisterCommand(&Command{})
}
