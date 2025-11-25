package reset

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct {
	soft  bool
	mixed bool
	hard  bool
}

func (c *Command) Name() string      { return "reset" }
func (c *Command) Aliases() []string { return []string{"drop"} }
func (c *Command) Usage() string     { return "reset [<commit-id>] [--soft|--mixed|--hard]" }
func (c *Command) Brief() string     { return "Reset current branch to a commit or HEAD" }
func (c *Command) Help() string {
	return `Reset current branch.

Options:
  --soft  : move HEAD only
  --mixed : move HEAD and reset index (default)
  --hard  : move HEAD, reset index and working directory

If <commit-id> is omitted, the last commit is used.

Usage:
  bvc reset [<commit-id>] [--soft|--mixed|--hard]

Examples:
  bvc reset
  bvc reset --mixed
  bvc reset --hard

  bvc reset <commit-id>
  bvc reset --soft <commit-id>
  bvc reset --mixed <commit-id>
  bvc reset --hard <commit-id>
`
}

func (c *Command) Subcommands() []command.Command { return nil }

func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.soft, "soft", false, "move HEAD only")
	fs.BoolVar(&c.mixed, "mixed", false, "move HEAD and reset index (default)")
	fs.BoolVar(&c.hard, "hard", false, "move HEAD, reset index and working directory")
}

func (c *Command) Run(ctx *command.Context) error {
	// determine mode
	mode := repo.ResetMixed
	count := 0
	if c.soft {
		mode = repo.ResetSoft
		count++
	}
	if c.mixed {
		mode = repo.ResetMixed
		count++
	}
	if c.hard {
		mode = repo.ResetHard
		count++
	}
	if count > 1 {
		return fmt.Errorf("conflicting reset modes specified")
	}

	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return err
	}

	commitID := ""
	if len(ctx.Args) > 0 {
		commitID = ctx.Args[0]
	}

	fmt.Printf("Resetting branch '%s' to commit %s (%s)...\n",
		func() string {
			b, _ := r.Meta.GetCurrentBranch()
			return b.Name
		}(), commitID, mode)

	if err := r.Reset(commitID, mode); err != nil {
		return err
	}

	fmt.Println("Reset complete.")
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
