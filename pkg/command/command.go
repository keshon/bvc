package command

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

type Registry struct {
	commands map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *Registry) Dispatch() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("no command specified")
	}

	args := os.Args[1:]
	return r.dispatchLevel(args, r.commands)
}

func (r *Registry) dispatchLevel(args []string, table map[string]Command) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	name := args[0]
	cmd, ok := table[name]
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	// if there are subcommands and the next arg matches them — dive deeper
	sub := cmd.SubCommands()
	if len(sub) > 0 && len(args) > 1 {
		subName := args[1]
		subTable := make(map[string]Command)
		for _, sc := range sub {
			subTable[sc.Name()] = sc
		}

		if _, exists := subTable[subName]; exists {
			return r.dispatchLevel(args[1:], subTable)
		}
	}

	// no subcommands left, this is final command — parse flags
	fs := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	cmd.Flags(fs)

	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}

	ctx := Context{
		Args: fs.Args(),
	}

	return cmd.Run(ctx)
}

func (r *Registry) Get(name string) Command {
	return r.commands[name]
}

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
