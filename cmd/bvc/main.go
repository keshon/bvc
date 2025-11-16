package main

import (
	"fmt"
	"os"

	"github.com/keshon/bvc/internal/command"

	// Register commands
	_ "github.com/keshon/bvc/internal/command/add"
	_ "github.com/keshon/bvc/internal/command/block"
	_ "github.com/keshon/bvc/internal/command/branch"
	_ "github.com/keshon/bvc/internal/command/checkout"
	_ "github.com/keshon/bvc/internal/command/cherry-pick"
	_ "github.com/keshon/bvc/internal/command/commit"
	_ "github.com/keshon/bvc/internal/command/help"
	_ "github.com/keshon/bvc/internal/command/init"
	_ "github.com/keshon/bvc/internal/command/log"
	_ "github.com/keshon/bvc/internal/command/merge"
	_ "github.com/keshon/bvc/internal/command/reset"
	_ "github.com/keshon/bvc/internal/command/status"
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
