package command

import "flag"

type Context struct {
	Args []string
}

type Command interface {
	Name() string
	Help() string
	Flags(fs *flag.FlagSet)
	Run(ctx Context) error
	SubCommands() []Command
}
