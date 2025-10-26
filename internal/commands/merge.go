package commands

import (
	"app/internal/cli"
	"app/internal/core"
	"app/internal/merge"
	"app/internal/middleware"
	"fmt"
)

type MergeCommand struct{}

func (c *MergeCommand) Name() string        { return "merge" }
func (c *MergeCommand) Usage() string       { return "merge <branch-name>" }
func (c *MergeCommand) Description() string { return "Merge another branch into current" }
func (c *MergeCommand) DetailedDescription() string {
	return "Merge another branch into current branch using three-way merge."
}

func (c *MergeCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}

	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	targetBranch := ctx.Args[0]
	if currentBranch == targetBranch {
		return fmt.Errorf("cannot merge branch into itself")
	}

	return merge.PerformMerge(currentBranch, targetBranch)
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&MergeCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
