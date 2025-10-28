package branch_name

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
)

// Command creates a new branch from the current branch
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "branch" }

// Git-style usage
func (c *Command) Usage() string { return "branch <branch-name>" }

func (c *Command) Description() string {
	return "Create a new branch from the current branch"
}

func (c *Command) DetailedDescription() string {
	return "Create a new branch from the current branch. Equivalent to 'git branch <name>'."
}

// Short alias and one-letter shortcut
func (c *Command) Aliases() []string { return []string{"nb"} }
func (c *Command) Short() string     { return "n" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	name := ctx.Args[0]

	branch, err := core.CreateBranch(name)
	if err != nil {
		return fmt.Errorf("failed to create branch '%s': %w", name, err)
	}

	fmt.Printf("Branch '%s' created successfully.\n", branch.Name)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&Command{}, middleware.WithBlockIntegrityCheck()),
	)
}
