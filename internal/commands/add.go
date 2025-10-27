package commands

import (
	"fmt"

	"app/internal/cli"
	"app/internal/storage/file"
)

// AddCommand implements Git-like 'add'
type AddCommand struct{}

func (c *AddCommand) Name() string  { return "add" }
func (c *AddCommand) Usage() string { return "add <file|dir|.>" }
func (c *AddCommand) Description() string {
	return "Stage files or directories for the next commit"
}
func (c *AddCommand) DetailedDescription() string {
	return `Stage changes for commit.
Use 'add .' to stage all files in the repository.`
}
func (c *AddCommand) Aliases() []string { return nil }
func (c *AddCommand) Short() string     { return "a" }

func (c *AddCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("no files specified")
	}

	var paths []string
	for _, arg := range ctx.Args {
		if arg == "." {
			all, err := file.ListAll()
			if err != nil {
				return err
			}
			paths = append(paths, all...)
		} else {
			paths = append(paths, arg)
		}
	}

	entries, err := file.BuildAll(paths)
	if err != nil {
		return err
	}

	if err := file.StageFiles(entries); err != nil {
		return err
	}

	fmt.Printf("Staged %d file(s)\n", len(entries))
	return nil
}

func init() {
	cli.RegisterCommand(&AddCommand{})
}
