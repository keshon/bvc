package commands

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/merge"
	"app/internal/middleware"
)

// MergeCommand merges another branch into the current branch
type MergeCommand struct{}

// Canonical name
func (c *MergeCommand) Name() string { return "merge" }

// Usage string
func (c *MergeCommand) Usage() string { return "merge <branch-name>" }

// Short description
func (c *MergeCommand) Description() string {
	return "Merge another branch into the current branch"
}

// Detailed description
func (c *MergeCommand) DetailedDescription() string {
	return `Perform a three-way merge of the specified branch into the current branch.
Conflicts may need manual resolution.`
}

// Optional aliases
func (c *MergeCommand) Aliases() []string { return []string{"mg"} }

// One-letter shortcut
func (c *MergeCommand) Short() string { return "M" }

// Run executes the merge
func (c *MergeCommand) Run(ctx *cli.Context) error {
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
	return merge.PerformMerge(currentBranch.Name, targetBranch)
}

// Register command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&MergeCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
