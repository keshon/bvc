package branch

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
)

// Command lists all branches and highlights the current one
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "branch" }

// Usage string
func (c *Command) Usage() string { return "branch" }

// Short description
func (c *Command) Description() string {
	return "List all branches"
}

// Detailed description
func (c *Command) DetailedDescription() string {
	return `List all branches in the repository.
The current branch is highlighted with '*'.`
}

// Optional aliases
func (c *Command) Aliases() []string { return []string{"br"} }

// One-letter shortcut
func (c *Command) Short() string { return "B" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	allBranches, err := core.Branches()
	if err != nil {
		return err
	}

	// Print branches
	for _, branch := range allBranches {
		prefix := "  "
		if branch.Name == currentBranch.Name {
			prefix = "* "
		}
		fmt.Println(prefix + branch.Name)
	}

	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(&Command{})
}
