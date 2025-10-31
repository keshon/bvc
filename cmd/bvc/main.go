package main

import (
	"fmt"
	"os"

	"app/internal/command"
	_ "app/internal/command/add"
	_ "app/internal/command/analyze"
	_ "app/internal/command/blocks"
	_ "app/internal/command/branch"
	_ "app/internal/command/checkout"
	_ "app/internal/command/cherry-pick"
	_ "app/internal/command/commit"
	_ "app/internal/command/help"
	_ "app/internal/command/init"
	_ "app/internal/command/log"
	_ "app/internal/command/merge"
	_ "app/internal/command/reset"
	_ "app/internal/command/status"
	_ "app/internal/command/verify"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bvc <command> [args...]")
		fmt.Println("Available commands:")
		for _, cmd := range command.AllCommands() {
			fmt.Printf("  %s: %s\n", cmd.Name(), cmd.Help())
		}
		os.Exit(0)
	}

	cmdName := os.Args[1]
	cmd, ok := command.GetCommand(cmdName)
	if !ok {
		fmt.Printf("Unknown command: %s\n", cmdName)
		os.Exit(1)
	}

	ctx := &command.Context{
		Args: os.Args[2:],
	}

	if err := cmd.Run(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
