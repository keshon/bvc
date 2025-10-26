package commands

import (
	"app/internal/cli"
	"app/internal/core"
	"fmt"
)

type ListBrancheCommand struct{}

func (c *ListBrancheCommand) Name() string        { return "list" }
func (c *ListBrancheCommand) Usage() string       { return "list" }
func (c *ListBrancheCommand) Description() string { return "List all branches" }
func (c *ListBrancheCommand) DetailedDescription() string {
	return "List branches and highlight current branch"
}

func (c *ListBrancheCommand) Run(ctx *cli.Context) error {
	curr, _ := core.CurrentBranch()
	names, err := core.Branches()
	if err != nil {
		return err
	}
	for _, name := range names {
		prefix := "  "
		if name == curr {
			prefix = "* "
		}
		fmt.Println(prefix + name)
	}
	return nil
}

func init() {
	cli.RegisterCommand(&ListBrancheCommand{})
}
