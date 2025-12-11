// Package command provides a lightweight registry and dispatcher for modular
// CLI commands. A Command registers itself at init-time and can define flags,
// help text, and optional subcommands. The registry supports both direct
// invocation (DispatchArgs) and automatic CLI invocation (DispatchCLI).
//
// Typical usage:
//
//	type HelloCommand struct{}
//
//	func (c *HelloCommand) Name() string  { return "hello" }
//	func (c *HelloCommand) Brief() string { return "Print greeting" }
//	func (c *HelloCommand) Flags(fs *flag.FlagSet) {}
//	func (c *HelloCommand) Run(ctx *command.Context) error {
//	    fmt.Println("Hello")
//	    return nil
//	}
//
//	func init() {
//	    command.DefaultRegistry.Register(&HelloCommand{})
//	}
//
// In main:
//
//	func main() {
//	    if err := command.DefaultRegistry.DispatchCLI(); err != nil {
//	        fmt.Println("Error:", err)
//	        os.Exit(1)
//	    }
//	}
//
// Commands may implement SubCommands() to provide nested structures like:
//
//	git add
//	git push
//
// Subcommands are optional. Flags are parsed per command using the standard
// "flag" package. The Context passed to Run includes the original context.Context,
// the flag.FlagSet, and the remaining positional arguments.
package command

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
)

// DefaultRegistry is a shared global registry for automatic command registration.
// Most applications will use it directly. For more advanced scenarios, create a
// custom Registry instance with NewRegistry().
var DefaultRegistry = NewRegistry()

// Registry holds a set of commands indexed by name. It is not inherently
// thread-safe but in typical CLI usage commands register only at startup.
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new empty command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry. Usually called from package init().
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// DispatchArgs executes the named command with the provided arguments.
// The caller provides a context for cancellation and timeouts.
func (r *Registry) DispatchArgs(ctx context.Context, cmdName string, args []string) error {
	return r.dispatchRecursive(ctx, cmdName, args, r.commands)
}

// DispatchCLI executes a command based on os.Args. This is a convenience
// method for main() and should be avoided when dispatching programmatically.
func (r *Registry) DispatchCLI() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("no command specified")
	}
	cmdName := os.Args[1]
	args := os.Args[2:]
	return r.DispatchArgs(context.Background(), cmdName, args)
}

// dispatchRecursive resolves nested subcommands, builds a FlagSet, parses flags,
// constructs the Context, and executes the command.
//
// Resolution algorithm:
//
//  1. look up current command by name in table
//  2. if next arg matches a subcommand, recurse into that subcommand
//  3. otherwise parse flags for current command
//  4. call Run with remaining positional args
func (r *Registry) dispatchRecursive(ctx context.Context, name string, args []string, table map[string]Command) error {
	cmd, ok := table[name]
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	// Check for subcommands
	for _, sub := range cmd.SubCommands() {
		if len(args) > 0 && args[0] == sub.Name() {
			return r.dispatchRecursive(ctx, sub.Name(), args[1:], tableFromCommands(cmd.SubCommands()))
		}
	}

	fs := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	cmd.Flags(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}

	c := &Context{
		Ctx:   ctx,
		Args:  fs.Args(),
		Flags: fs,
	}

	return cmd.Run(c)
}

// tableFromCommands converts a slice of Command into a lookup table.
func tableFromCommands(cmds []Command) map[string]Command {
	m := make(map[string]Command, len(cmds))
	for _, c := range cmds {
		m[c.Name()] = c
	}
	return m
}

// Get returns a registered command by name or nil if missing.
func (r *Registry) Get(name string) Command {
	return r.commands[name]
}

// GetAll returns all commands sorted lexicographically by name.
// Useful for "help" commands and auto-completion.
func (r *Registry) GetAll() []Command {
	list := make([]Command, 0, len(r.commands))
	for _, c := range r.commands {
		list = append(list, c)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}
