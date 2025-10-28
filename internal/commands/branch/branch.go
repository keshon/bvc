package branch

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
)

// Command implements `branch` listing and creation (Git-style)
type Command struct{}

func (c *Command) Name() string  { return "branch" }
func (c *Command) Usage() string { return "branch [<branch-name>]" }
func (c *Command) Description() string {
	return "List all branches or create a new one"
}
func (c *Command) DetailedDescription() string {
	return `Usage:
  branch             - List all branches (current marked with '*')
  branch <name>      - Create a new branch from the current one`
}
func (c *Command) Aliases() []string { return []string{"br"} }
func (c *Command) Short() string     { return "B" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	// If there’s an argument — create new branch
	if len(ctx.Args) > 0 {
		name := ctx.Args[0]
		branch, err := core.CreateBranch(name)
		if err != nil {
			return fmt.Errorf("failed to create branch '%s': %w", name, err)
		}
		fmt.Printf("Branch '%s' created successfully.\n", branch.Name)
		return nil
	}

	// Otherwise — list branches
	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	allBranches, err := core.Branches()
	if err != nil {
		return err
	}

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
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&Command{}, middleware.WithBlockIntegrityCheck()),
	)
}
