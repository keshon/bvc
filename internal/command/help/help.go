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
		// pass all arguments joined with space
		return runCommandHelp(strings.Join(ctx.Args, " "))
	}
	return runListAllCommands()
}

// runCommandHelp shows detailed help for a specific command
func runCommandHelp(fullPath string) error {
	args := strings.Split(fullPath, " ")

	// keywords and their ANSI color
	var helpKeywords = map[string]string{
		"Usage:":                 "\033[32m",
		"Options:":               "\033[32m",
		"Examples:":              "\033[32m",
		"Available subcommands:": "\033[32m",
	}

	// reset ANSI code
	const ansiReset = "\033[0m"

	// resolve full path in command tree
	node, remaining, err := command.ResolveCommand(args)
	if err != nil || node == nil || node.Cmd == nil || len(remaining) > 0 {
		fmt.Printf("Unknown command: %s\n", fullPath)
		return nil
	}

	cmd := node.Cmd

	// get raw help string
	helpText := cmd.Help()

	// recolor keywords on the fly
	for k, style := range helpKeywords {
		helpText = strings.ReplaceAll(helpText, k, style+k+ansiReset)
	}

	// print recolored help
	fmt.Println(helpText)

	// if the command has subcommands, show them
	subcmds := cmd.Subcommands()
	if len(subcmds) > 0 {
		fmt.Println("\nSubcommands:")

		// compute longest name for alignment
		longest := 0
		for _, sc := range subcmds {
			if l := len(sc.Name()); l > longest {
				longest = l
			}
		}

		for _, sc := range subcmds {
			desc := sc.Brief()
			if desc == "" {
				desc = "-"
			}
			padding := strings.Repeat(" ", longest-len(sc.Name())+2)
			fmt.Printf("  \033[1m%s\033[0m%s%s\n", sc.Name(), padding, desc)
		}

		fmt.Printf("\nType 'help %s <subcommand>' for detailed info on a subcommand.\n", fullPath)
	}

	return nil
}

// runListAllCommands lists all leaf commands (exclude parents without children)
func runListAllCommands() error {
	commands := command.AllCommands()
	type CmdEntry struct {
		Path string
		Cmd  command.Command
	}

	var entries []CmdEntry

	var walk func(prefix string, cmd command.Command)
	walk = func(prefix string, cmd command.Command) {
		path := cmd.Name()
		if prefix != "" {
			path = prefix + " " + cmd.Name()
		}

		// only add leaf commands (no subcommands)
		if len(cmd.Subcommands()) == 0 {
			entries = append(entries, CmdEntry{Path: path, Cmd: cmd})
		}

		for _, sc := range cmd.Subcommands() {
			walk(path, sc)
		}
	}

	// start from root-level commands
	for _, cmd := range commands {
		walk("", cmd)
	}

	// sort alphabetically by path
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	// find longest path for padding
	longest := 0
	for _, e := range entries {
		if l := len(e.Path); l > longest {
			longest = l
		}
	}

	fmt.Print("Available commands:\n\n")
	for _, e := range entries {
		desc := e.Cmd.Brief()
		if desc == "" {
			desc = "-"
		}
		padding := strings.Repeat(" ", longest-len(e.Path)+2)
		fmt.Printf("  \033[1m%s\033[0m%s%s\n", e.Path, padding, desc)
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
