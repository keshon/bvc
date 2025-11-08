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
func (c *Command) Aliases() []string { return []string{"lg"} }
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
  bvc log
  bvc log -a
  bvc log --oneline -n 10
  bvc log main
`
}

func (c *Command) Run(ctx *command.Context) error {
	fs := flag.NewFlagSet("log", flag.ContinueOnError)

	showAll := fs.Bool("all", false, "show commits from all branches")
	fs.BoolVar(showAll, "a", false, "alias for --all")

	oneline := fs.Bool("oneline", false, "show each commit on one line")

	n := fs.Int("n", 0, "limit number of commits")
	since := fs.String("since", "", "show commits after date YYYY-MM-DD")
	until := fs.String("until", "", "show commits before date YYYY-MM-DD")

	if err := fs.Parse(ctx.Args); err != nil {
		return err
	}

	branchArg := ""
	args := fs.Args()
	if len(args) > 0 {
		branchArg = args[0]
	}

	return c.log(*showAll, *oneline, *n, *since, *until, branchArg)
}

func (c *Command) log(showAll, oneline bool, n int, since, until, branchArg string) error {
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

			// Filter by date
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

	// Sort newest first
	sort.Slice(commits, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, commits[i].Timestamp)
		tj, _ := time.Parse(time.RFC3339, commits[j].Timestamp)
		return ti.After(tj)
	})

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
