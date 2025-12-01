package add

import (
	"flag"
	"fmt"
	"strings"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct {
	all    bool
	update bool
}

func (c *Command) Name() string      { return "add" }
func (c *Command) Aliases() []string { return nil }
func (c *Command) Brief() string     { return "Stage files or directories for the next commit" }
func (c *Command) Usage() string     { return "add <file|dir|.> [options]" }
func (c *Command) Help() string {
	return `Stage changes for commit.

Options:
  -a, --all             Stage all changes, including deletions (-A)
   --update             Stage modifications and deletions only (-u)

Usage:
  bvc add <file|dir|.> [options]

Examples:
  bvc add .
  bvc add 'main.go'
  bvc add dir/
`
}
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.all, "all", false, "Stage all changes, including deletions (-A)")
	fs.BoolVar(&c.update, "update", false, "Stage modifications and deletions only (-u)")
}

func (c *Command) Run(ctx *command.Context) error {
	args := filterNonFlags(ctx.Args)

	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	entries, err := r.Add(args, c.all, c.update)
	if err != nil {
		return err
	}

	fmt.Printf("Staged %d file(s)\n", len(entries))
	return nil
}

// helper: remove flags from args
func filterNonFlags(args []string) []string {
	var res []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		res = append(res, a)
	}
	return res
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithBlockIntegrityCheck(),
			middleware.WithDebugArgsPrint(),
		),
	)
}
