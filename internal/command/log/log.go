package log

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct {
	all     bool
	oneline bool
	limit   int
	since   string
	until   string
}

func (c *Command) Name() string      { return "log" }
func (c *Command) Aliases() []string { return []string{"commits"} }
func (c *Command) Usage() string     { return "log [options] [branch]" }
func (c *Command) Brief() string     { return "Show commit history (current branch by default)" }
func (c *Command) Help() string {
	return `Show commit logs.

Options:
  -a, --all             Show commits from all branches.
      --oneline         Show each commit as a single line (ID + message).
  -n <count>            Limit to the last N commits.
      --since <date>    Show commits after the given date (YYYY-MM-DD).
      --until <date>    Show commits before the given date (YYYY-MM-DD).

Usage:
  bvc log [options]

Examples:
  bvc log
  bvc log -a
  bvc log --oneline -n 10
  bvc log main
`
}

func (c *Command) Subcommands() []command.Command {
	return nil
}
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.all, "all", false, "show commits from all branches")
	fs.BoolVar(&c.all, "a", false, "alias for --all")

	fs.BoolVar(&c.oneline, "oneline", false, "show each commit on one line")

	fs.IntVar(&c.limit, "n", 0, "limit number of commits")

	fs.StringVar(&c.since, "since", "", "show commits after date YYYY-MM-DD")

	fs.StringVar(&c.until, "until", "", "show commits before date YYYY-MM-DD")
}

func (c *Command) Run(ctx *command.Context) error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	var since, until *time.Time
	if c.since != "" {
		t, _ := time.Parse("2006-01-02", c.since)
		since = &t
	}
	if c.until != "" {
		t, _ := time.Parse("2006-01-02", c.until)
		until = &t
	}

	branch := ""
	args := ctx.Flags.Args()
	if len(args) > 0 {
		branch = args[0]
	}

	commits, err := r.Log(repo.CommitFilter{
		AllBranches: c.all,
		Branch:      branch,
		Limit:       c.limit,
		Since:       since,
		Until:       until,
	})
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		fmt.Println("No commits found")
		return nil
	}

	for _, cmt := range commits {
		if c.oneline {
			short := cmt.Commit.ID[:7]
			msg := strings.SplitN(cmt.Commit.Message, "\n", 2)[0]
			if len(cmt.Refs) > 0 {
				fmt.Printf("%s (%s) %s\n", short, strings.Join(cmt.Refs, ", "), msg)
			} else {
				fmt.Printf("%s %s\n", short, msg)
			}
		} else {
			t, _ := time.Parse(time.RFC3339, cmt.Commit.Timestamp)
			fmt.Printf("\033[33mcommit\033[0m %s", cmt.Commit.ID)
			if len(cmt.Refs) > 0 {
				fmt.Printf(" (%s)", strings.Join(cmt.Refs, ", "))
			}
			fmt.Println()
			if len(cmt.Commit.Parents) > 1 {
				fmt.Printf("Merge: %s\n", strings.Join(cmt.Commit.Parents, " "))
			}
			fmt.Printf("Date:   %s\n\n", t.Format("Mon Jan 2 15:04:05 2006 -0700"))
			for _, line := range strings.Split(cmt.Commit.Message, "\n") {
				if strings.TrimSpace(line) == "" {
					fmt.Println()
				} else {
					fmt.Printf("    %s\n", line)
				}
			}
			fmt.Println()
		}
	}

	return nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
