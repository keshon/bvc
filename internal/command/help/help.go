package help

import (
	"app/internal/command"
	"app/internal/middleware"
	"fmt"
	"sort"
	"strings"
)

type Command struct{}

func (c *Command) Name() string      { return "help" }
func (c *Command) Short() string     { return "H" }
func (c *Command) Aliases() []string { return []string{"h", "?"} }
func (c *Command) Usage() string     { return "help [command]" }
func (c *Command) Brief() string     { return "Show help for commands" }
func (c *Command) Help() string {
	return "Display detailed help information for a specific command, or list all commands if none is provided."
}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) > 0 {
		return commandHelp(strings.ToLower(ctx.Args[0]))
	}
	return listAllCommands()
}

func commandHelp(name string) error {
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
		aliases := aliasesCmd.Aliases()
		if len(aliases) > 0 {
			fmt.Printf("Aliases: %s\n", strings.Join(aliases, ", "))
		}
	}

	return nil
}

func listAllCommands() error {
	commands := command.AllCommands()
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
		desc := cmd.Brief()
		if desc == "" {
			desc = "-"
		}

		fmt.Printf("%-10s \033[90m%-54s\033[0m %-32s\n", cmd.Name(), desc, usage)
	}

	fmt.Println("\nUse 'help <command>' to see detailed usage of a specific command.")
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
