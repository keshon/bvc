package cherry_pick

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct{}

func (c *Command) Name() string  { return "cherry-pick" }
func (c *Command) Brief() string { return "Apply selected commit to the current branch" }
func (c *Command) Usage() string { return "cherry-pick <commit-id>" }
func (c *Command) Help() string {
	return `Apply a specific commit to the current branch.

Usage:
  cherry-pick <commit-id>`
}
func (c *Command) Aliases() []string              { return []string{"cp"} }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet)         {}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit ID required")
	}

	commitID := ctx.Args[0]

	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	newCommit, err := r.CherryPick(commitID)
	if err != nil {
		return err
	}

	fmt.Printf("Picked %s to %s\n", commitID, newCommit.ID)
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
