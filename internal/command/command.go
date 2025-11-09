package command

import (
	"flag"
)

// Command represents a single CLI command or subcommand.
type Command interface {
	Name() string
	Brief() string
	Usage() string
	Help() string
	Aliases() []string
	Subcommands() []Command
	Flags(fs *flag.FlagSet)
	Run(ctx *Context) error
}

// Context holds runtime info for a command execution.
type Context struct {
	Args  []string
	Flags *flag.FlagSet
}
