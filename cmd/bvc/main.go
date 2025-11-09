package main

import (
	"fmt"
	"os"

	"app/internal/command"

	// Register commands
	_ "app/internal/command/add"
	_ "app/internal/command/block"
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
)

func main() {
	args := os.Args[1:]

	// No arguments? Print usage and exit
	if len(args) == 0 {
		fmt.Println("Usage: bvc <command> [args...]")
		fmt.Println("Available commands:")
		for _, cmd := range command.AllCommands() {
			fmt.Printf("  %s: %s\n", cmd.Name(), cmd.Brief())
		}
		os.Exit(0)
	}

	// Delegate to runner
	command.RunCLI(args)
}
