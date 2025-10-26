package commands

import (
	"app/internal/cli"
	"app/internal/core"
	"fmt"
)

type NewBranchCommand struct{}

func (c *NewBranchCommand) Name() string        { return "new" }
func (c *NewBranchCommand) Usage() string       { return "new <branch-name>" }
func (c *NewBranchCommand) Description() string { return "Create a new branch" }
func (c *NewBranchCommand) DetailedDescription() string {
	return "Create a new branch from the current branch"
}

func (c *NewBranchCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	name := ctx.Args[0]
	return core.CreateBranch(name)
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&NewBranchCommand{}, cli.WithRepoCheck()),
	)
}
