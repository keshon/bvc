package branch

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
)

type Command struct{}

func (c *Command) Name() string      { return "branch" }
func (c *Command) Short() string     { return "B" }
func (c *Command) Aliases() []string { return []string{"br"} }
func (c *Command) Usage() string     { return "branch [<branch-name>]" }
func (c *Command) Brief() string     { return "List all branches or create a new one" }

func (c *Command) Help() string {
	return `List all branches or create a new one.

Usage:
  branch        - list all branches (current marked with '*')
  branch <name> - create a new branch from the current one`
}

func (c *Command) Run(ctx *command.Context) error {
	// open the repository context
	repo, err := repo.OpenAt(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// case 1: create new branch
	if len(ctx.Args) > 0 {
		name := ctx.Args[0]
		newBranch, err := repo.CreateBranch(name)
		if err != nil {
			return fmt.Errorf("failed to create branch %q: %w", name, err)
		}
		fmt.Printf("Branch '%s' created successfully.\n", newBranch.Name)
		return nil
	}

	// case 2: list branches
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to determine current branch: %w", err)
	}

	// list all branches
	fmt.Println("Branches:")
	allBranches, err := repo.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
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

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
			middleware.WithBlockIntegrityCheck(),
		),
	)
}
