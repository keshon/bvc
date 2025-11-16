package branch

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct{}

func (c *Command) Name() string  { return "branch" }
func (c *Command) Brief() string { return "List all branches or create a new one" }
func (c *Command) Usage() string { return "branch [options] [<branch-name>]" }
func (c *Command) Help() string {
	return `List all branches or create a new one.

Usage:
  branch           - list all branches (current marked with '*')
  branch <name>    - create a new branch from the current one`
}
func (c *Command) Aliases() []string              { return []string{"br", "B"} }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet)         {}

func (c *Command) Run(ctx *command.Context) error {
	// Open the repository
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	args := ctx.Args

	// If a branch name is given, create a new branch
	if len(args) > 0 {
		name := args[0]
		newBranch, err := r.Meta.CreateBranch(name)
		if err != nil {
			return fmt.Errorf("failed to create branch %q: %w", name, err)
		}
		fmt.Printf("Branch '%s' created successfully.\n", newBranch.Name)
		return nil
	}

	// Otherwise list all branches
	current, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	allBranches, err := r.Meta.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	fmt.Println("Branches:")
	for _, b := range allBranches {
		prefix := "  "
		if b.Name == current.Name {
			prefix = "* "
		}
		fmt.Println(prefix + b.Name)
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
