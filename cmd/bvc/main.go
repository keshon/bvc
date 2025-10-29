package main

import (
	"fmt"
	"os"

	"app/internal/cli"
	_ "app/internal/commands"
	"app/internal/core"
)

func main() {
	if err := core.InitRepo(); err != nil {
		fmt.Printf("Failed to initialize Binary Version Control dirs: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: bvc <command> [args...]")
		fmt.Println("Available commands:")
		for _, cmd := range cli.AllCommands() {
			fmt.Printf("  %s: %s\n", cmd.Name(), cmd.Help())
		}
		os.Exit(0)
	}

	cmdName := os.Args[1]
	cmd, ok := cli.GetCommand(cmdName)
	if !ok {
		fmt.Printf("Unknown command: %s\n", cmdName)
		os.Exit(1)
	}

	ctx := &cli.Context{
		Args: os.Args[2:],
	}

	if err := cmd.Run(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
