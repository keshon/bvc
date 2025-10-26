package commands

import (
	"fmt"
	"sort"
	"strings"

	"app/internal/cli"
)

type HelpCommand struct{}

func (c *HelpCommand) Name() string        { return "help" }
func (c *HelpCommand) Usage() string       { return "help <command-name>" }
func (c *HelpCommand) Description() string { return "Show help for commands" }
func (c *HelpCommand) DetailedDescription() string {
	return "Show help information for a specific command."
}

func (c *HelpCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) > 0 {
		return commandHelp(ctx.Args[0])
	}
	return allCommands()
}

func commandHelp(name string) error {
	cmd, ok := cli.GetCommand(name)
	if !ok {
		fmt.Printf("Unknown command: %s\n", name)
		return nil
	}

	if u := cmd.Usage(); u != "" {
		fmt.Printf("\033[90mUsage:\033[0m %s\n\n", u)
	}
	fmt.Printf("%s\n", cmd.DetailedDescription())
	fmt.Println()
	return nil
}

func allCommands() error {
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

func init() {
	cli.RegisterCommand(&HelpCommand{})
}
