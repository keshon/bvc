package branch

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct {
	delete      bool
	forceDelete bool
	move        bool
	forceMove   bool
	forceCreate bool
}

func (c *Command) Name() string  { return "branch" }
func (c *Command) Brief() string { return "List all branches or create a new one" }
func (c *Command) Usage() string { return "branch [options] [<branch-name>]" }
func (c *Command) Help() string {
	return `Create, list, rename, or delete branches.

Options:
  -d               Delete branch (safe; refuses if branch is not fully merged)
  -D               Delete branch (force; deletes even if unmerged)
  -m               Rename branch
  -M               Rename branch (force; overwrite destination)

Usage:
  bvc branch                     List all branches (current marked with '*')
  bvc branch <name>              Create a new branch at the current HEAD
  bvc branch <name> <start>      Create a new branch at the given commit/branch
  bvc branch -d <name>           Delete a branch
  bvc branch -D <name>           Force-delete a branch
  bvc branch -m <old> <new>      Rename a branch
  bvc branch -M <old> <new>      Force-rename (overwrite if exists)

Examples:
  bvc branch
  bvc branch feature
  bvc branch feature main
  bvc branch -d feature
  bvc branch -m old-name new-name
`
}

func (c *Command) Aliases() []string              { return []string{"br", "B"} }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.delete, "d", false, "delete branch (safe)")
	fs.BoolVar(&c.forceDelete, "D", false, "delete branch (force)")
	fs.BoolVar(&c.move, "m", false, "rename branch")
	fs.BoolVar(&c.forceMove, "M", false, "rename branch (force)")
	fs.BoolVar(&c.forceCreate, "f", false, "create or reset branch forcefully")
}

func (c *Command) Run(ctx *command.Context) error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	args := ctx.Args

	// delete
	if c.delete || c.forceDelete {
		if len(args) != 1 {
			return fmt.Errorf("usage: branch -d|-D <name>")
		}
		return r.DeleteBranch(args[0], c.forceDelete)
	}

	// rename
	if c.move || c.forceMove {
		if len(args) != 2 {
			return fmt.Errorf("usage: branch -m|-M <old> <new>")
		}
		return r.RenameBranch(args[0], args[1], c.forceMove)
	}

	// create
	if len(args) == 1 || len(args) == 2 {
		name := args[0]
		start := ""
		if len(args) == 2 {
			start = args[1]
		}

		_, err := r.CreateBranch(name, start, c.forceCreate)
		return err
	}

	// list
	current, branches, err := r.ListBranches()
	if err != nil {
		return err
	}

	for _, b := range branches {
		if b == current {
			fmt.Printf("* %s\n", b)
		} else {
			fmt.Printf("  %s\n", b)
		}
	}
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
