package commands

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
)

// BranchCommand lists all branches and highlights the current one
type BranchCommand struct{}

// Canonical name
func (c *BranchCommand) Name() string { return "branch" }

// Usage string
func (c *BranchCommand) Usage() string { return "branch" }

// Short description
func (c *BranchCommand) Description() string {
	return "List all branches"
}

// Detailed description
func (c *BranchCommand) DetailedDescription() string {
	return `List all branches in the repository.
The current branch is highlighted with '*'.`
}

// Optional aliases
func (c *BranchCommand) Aliases() []string { return []string{"br"} }

// One-letter shortcut
func (c *BranchCommand) Short() string { return "B" }

// Run executes the command
func (c *BranchCommand) Run(ctx *cli.Context) error {
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
			prefix = "* " // Highlight current branch
		}
		fmt.Println(prefix + branch.Name)
	}

	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(&BranchCommand{})
}
