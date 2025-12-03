package checkout

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct{}

func (c *Command) Name() string  { return "checkout" }
func (c *Command) Brief() string { return "Switch to another branch" }
func (c *Command) Usage() string { return "checkout <branch-name>" }
func (c *Command) Help() string {
	return `Switch to another branch.

Usage:
  checkout <branch-name>`
}
func (c *Command) Aliases() []string              { return []string{"co"} }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet)         {}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	branchName := ctx.Args[0]

	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	result, err := r.Checkout(branchName)
	if err != nil {
		return err
	}

	if result.Empty {
		fmt.Printf("Branch '%s' is empty — switched.\n", result.Branch)
		return nil
	}

	fmt.Printf("Switched to branch '%s' at commit %s\n", result.Branch, result.CommitID)
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
