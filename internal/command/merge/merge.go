package merge

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
)

type Command struct{}

func (c *Command) Name() string      { return "merge" }
func (c *Command) Short() string     { return "M" }
func (c *Command) Aliases() []string { return []string{"mg"} }
func (c *Command) Usage() string     { return "merge <branch-name>" }
func (c *Command) Brief() string {
	return "Merge another branch into the current branch"
}
func (c *Command) Help() string {
	return `Perform a three-way merge of the specified branch into the current branch.
Conflicts may need manual resolution.`
}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}

	// Open the repository context
	r, err := repo.OpenAt(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	GetCurrentBranch, err := r.GetCurrentBranch()
	if err != nil {
		return err
	}

	targetBranch := ctx.Args[0]
	if GetCurrentBranch.Name == targetBranch {
		return fmt.Errorf("cannot merge branch into itself")
	}

	fmt.Printf("Merging branch '%s' into '%s'...\n", targetBranch, GetCurrentBranch.Name)
	return merge(GetCurrentBranch.Name, targetBranch)
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
