package merge

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct{}

func (c *Command) Name() string      { return "merge" }
func (c *Command) Aliases() []string { return []string{"mg"} }
func (c *Command) Usage() string     { return "merge <branch-name>" }
func (c *Command) Brief() string     { return "Merge another branch into the current branch" }
func (c *Command) Help() string {
	return `Perform a three-way merge of the specified branch into the current branch.
Conflicts may need manual resolution.`
}
func (c *Command) Subcommands() []command.Command {
	return nil
}
func (c *Command) Flags(fs *flag.FlagSet) {}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	targetBranch := ctx.Args[0]

	// Open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	currentBranch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	if currentBranch.Name == targetBranch {
		return fmt.Errorf("cannot merge branch into itself")
	}

	fmt.Printf("Merging branch '%s' into '%s'...\n", targetBranch, currentBranch.Name)
	return merge(currentBranch.Name, targetBranch)
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
