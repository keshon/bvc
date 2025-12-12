package command

import (
	"context"
	"flag"
)

type Context struct {
	Ctx   context.Context
	Args  []string
	Flags *flag.FlagSet
}

type Command interface {
	Name() string
	Brief() string
	Help() string
	Flags(fs *flag.FlagSet)
	Run(ctx *Context) error
	SubCommands() []Command
}
