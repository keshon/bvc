package merge

import (
	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
	"fmt"
)

// Command merges another branch into the current branch
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "merge" }

// Usage string
func (c *Command) Usage() string { return "merge <branch-name>" }

// Short description
func (c *Command) Description() string {
	return "Merge another branch into the current branch"
}

// Detailed description
func (c *Command) DetailedDescription() string {
	return `Perform a three-way merge of the specified branch into the current branch.
Conflicts may need manual resolution.`
}

// Optional aliases
func (c *Command) Aliases() []string { return []string{"mg"} }

// One-letter shortcut
func (c *Command) Short() string { return "M" }

// Run executes the merge
func (c *Command) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}

	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	targetBranch := ctx.Args[0]
	if currentBranch.Name == targetBranch {
		return fmt.Errorf("cannot merge branch into itself")
	}

	fmt.Printf("Merging branch '%s' into '%s'...\n", targetBranch, currentBranch.Name)
	return performMerge(currentBranch.Name, targetBranch)
}

// Register command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&Command{}, middleware.WithBlockIntegrityCheck()),
	)
}
