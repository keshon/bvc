package branch

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/core"
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
	return `Usage:
  branch             - List all branches (current marked with '*')
  branch <name>      - Create a new branch from the current one`
}

func (c *Command) Run(ctx *command.Context) error {
	// Ensure repository is valid
	fmt.Println("Checking repository integrity...")
	if err := repo.VerifyBlocks(false); err != nil {
		return fmt.Errorf(
			"repository verification failed: %v\nPlease run `bvc repair` before continuing",
			err,
		)
	}

	// Open the current repository (using .bvc)
	r, err := core.OpenAt(config.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Case 1: Create new branch
	if len(ctx.Args) > 0 {
		name := ctx.Args[0]
		newBranch, err := r.CreateBranch(name)
		if err != nil {
			return fmt.Errorf("failed to create branch %q: %w", name, err)
		}
		fmt.Printf("Branch '%s' created successfully.\n", newBranch.Name)
		return nil
	}

	// Case 2: List branches
	current, err := r.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to determine current branch: %w", err)
	}

	all, err := r.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	for _, br := range all {
		prefix := "  "
		if br.Name == current.Name {
			prefix = "* "
		}
		fmt.Println(prefix + br.Name)
	}

	return nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
