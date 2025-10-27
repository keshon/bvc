package commands

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
)

// NewBranchCommand creates a new branch from the current branch
type NewBranchCommand struct{}

// Canonical name
func (c *NewBranchCommand) Name() string { return "branch" }

// Git-style usage
func (c *NewBranchCommand) Usage() string { return "branch <branch-name>" }

func (c *NewBranchCommand) Description() string {
	return "Create a new branch from the current branch"
}

func (c *NewBranchCommand) DetailedDescription() string {
	return "Create a new branch from the current branch. Equivalent to 'git branch <name>'."
}

// Short alias and one-letter shortcut
func (c *NewBranchCommand) Aliases() []string { return []string{"nb"} }
func (c *NewBranchCommand) Short() string     { return "n" }

// Run executes the command
func (c *NewBranchCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	name := ctx.Args[0]

	// Create the branch using core API
	_, err := core.CreateBranch(name)
	if err != nil {
		return fmt.Errorf("failed to create branch '%s': %w", name, err)
	}

	fmt.Printf("Branch '%s' created successfully.\n", name)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&NewBranchCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
