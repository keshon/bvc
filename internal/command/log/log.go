package log

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repo/meta"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Command struct{}

func (c *Command) Name() string      { return "log" }
func (c *Command) Short() string     { return "l" }
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
  bvc log main`
}

func (c *Command) Subcommands() []command.Command {
	return nil
}
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.Bool("all", false, "show commits from all branches")
	fs.Bool("a", false, "alias for --all")
	fs.Bool("oneline", false, "show each commit on one line")
	fs.Int("n", 0, "limit number of commits")
	fs.String("since", "", "show commits after date YYYY-MM-DD")
	fs.String("until", "", "show commits before date YYYY-MM-DD")
}

func (c *Command) Run(ctx *command.Context) error {
	showAll := ctx.Flags.Lookup("all").Value.(flag.Getter).Get().(bool) ||
		ctx.Flags.Lookup("a").Value.(flag.Getter).Get().(bool)
	oneline := ctx.Flags.Lookup("oneline").Value.(flag.Getter).Get().(bool)
	n := ctx.Flags.Lookup("n").Value.(flag.Getter).Get().(int)
	since := ctx.Flags.Lookup("since").Value.(flag.Getter).Get().(string)
	until := ctx.Flags.Lookup("until").Value.(flag.Getter).Get().(string)

	branchArg := ""
	args := ctx.Flags.Args()
	if len(args) > 0 {
		branchArg = args[0]
	}

	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	var branches []string
	if branchArg != "" {
		branches = []string{branchArg}
	} else if showAll {
		all, err := r.Meta.ListBranches()
		if err != nil {
			return fmt.Errorf("failed to list branches: %w", err)
		}
		for _, b := range all {
			branches = append(branches, b.Name)
		}
	} else {
		cur, err := r.Meta.GetCurrentBranch()
		if err != nil {
			return err
		}
		branches = []string{cur.Name}
	}

	var commits []*meta.Commit
	seen := make(map[string]bool)

	for _, branch := range branches {
		branchCommits, err := r.Meta.GetCommitsForBranch(branch)
		if err != nil {
			return fmt.Errorf("failed to get commits for branch %q: %w", branch, err)
		}

		for _, cmt := range branchCommits {
			if seen[cmt.ID] {
				continue
			}
			seen[cmt.ID] = true

			t, err := time.Parse(time.RFC3339, cmt.Timestamp)
			if err != nil {
				continue
			}
			if since != "" {
				s, _ := time.Parse("2006-01-02", since)
				if t.Before(s) {
					continue
				}
			}
			if until != "" {
				u, _ := time.Parse("2006-01-02", until)
				if t.After(u) {
					continue
				}
			}

			commits = append(commits, cmt)
		}
	}

	if len(commits) == 0 {
		fmt.Println("No commits found")
		return nil
	}

	sort.Slice(commits, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, commits[i].Timestamp)
		tj, _ := time.Parse(time.RFC3339, commits[j].Timestamp)
		return ti.After(tj)
	})

	if n > 0 && n < len(commits) {
		commits = commits[:n]
	}

	if oneline {
		for _, cmt := range commits {
			firstLine := strings.SplitN(cmt.Message, "\n", 2)[0]
			fmt.Printf("%s %s\n", cmt.ID, firstLine)
		}
	} else {
		for _, cmt := range commits {
			t, _ := time.Parse(time.RFC3339, cmt.Timestamp)
			fmt.Printf("\033[90mCommit:\033[0m %s\n", cmt.ID)
			fmt.Printf("\033[90mBranch:\033[0m %s\n", cmt.Branch)
			if len(cmt.Parents) > 0 {
				fmt.Printf("\033[90mParent:\033[0m %s\n", strings.Join(cmt.Parents, " "))
			}
			fmt.Printf("\033[90mDate:\033[0m   %s\n\n", t.Format("Mon Jan 2 15:04:05 2006"))

			lines := strings.Split(cmt.Message, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					fmt.Println()
				} else {
					fmt.Printf("    %s\n", line)
				}
			}
			fmt.Println()
		}
	}

	fmt.Printf("Total commits: %d\n", len(commits))
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
