package command

import (
	"flag"
	"fmt"
	"os"
)

// RunCLI is the main entrypoint for executing commands.
// It parses arguments, resolves subcommands, applies flags, and runs the target command.
func RunCLI(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no command provided")
		os.Exit(1)
	}

	node, remaining, err := ResolveCommand(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	cmd := node.Cmd

	fs := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	cmd.Flags(fs)
	if err := fs.Parse(remaining); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing flags:", err)
		os.Exit(1)
	}

	ctx := &Context{
		Args:  fs.Args(),
		Flags: fs,
	}

	if err := cmd.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
