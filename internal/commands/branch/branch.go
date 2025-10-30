package branch

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/repo"
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

func (c *Command) Run(ctx *cli.Context) error {
	// If there’s an argument — create new branch
	if len(ctx.Args) > 0 {
		fmt.Println("Checking repository integrity...")
		if err := repo.VerifyBlocks(false); err != nil {
			return fmt.Errorf(
				"repository verification failed: %v\nPlease run `bvc repair` before continuing",
				err,
			)
		}
		name := ctx.Args[0]
		branch, err := core.CreateBranch(name)
		if err != nil {
			return fmt.Errorf("failed to create branch '%s': %w", name, err)
		}
		fmt.Printf("Branch '%s' created successfully.\n", branch.Name)
		return nil
	}

	// Otherwise — list branches
	GetCurrentBranch, err := core.GetCurrentBranch()
	if err != nil {
		return err
	}

	allBranches, err := core.GetBranches()
	if err != nil {
		return err
	}

	for _, branch := range allBranches {
		prefix := "  "
		if branch.Name == GetCurrentBranch.Name {
			prefix = "* "
		}
		fmt.Println(prefix + branch.Name)
	}

	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
