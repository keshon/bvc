package log

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Command struct{}

func (c *Command) Name() string      { return "log" }
func (c *Command) Short() string     { return "l" }
func (c *Command) Aliases() []string { return []string{"lg"} }
func (c *Command) Usage() string     { return "log [-a|--all]" }
func (c *Command) Brief() string     { return "Show commit history (current branch by default)" }
func (c *Command) Help() string {
	return "List commits for the current branch or all branches if -a / --all is specified."
}

func (c *Command) Run(ctx *command.Context) error {
	showAll := false
	for _, arg := range ctx.Args {
		if arg == "--all" {
			showAll = true
		}
	}

	return c.log(showAll)
}

func (c *Command) log(showAll bool) error {
	// Open the repository context
	r, err := repo.OpenAt(config.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	currentBranch, err := r.GetCurrentBranch()
	if err != nil {
		return err
	}

	var branchNames []string
	if showAll {
		allBranches, err := r.ListBranches()
		if err != nil {
			return fmt.Errorf("failed to get branches: %w", err)
		}
		for _, b := range allBranches {
			branchNames = append(branchNames, b.Name)
		}
	} else {
		branchNames = []string{currentBranch.Name}
	}

	var commits []*repo.Commit
	seen := make(map[string]bool)

	for _, branch := range branchNames {
		branchCommits, err := r.GetCommitsForBranch(branch)
		if err != nil {
			return fmt.Errorf("failed to get commits for branch %q: %w", branch, err)
		}

		for _, cmt := range branchCommits {
			if !seen[cmt.ID] {
				seen[cmt.ID] = true
				commits = append(commits, cmt)
			}
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

	if showAll {
		fmt.Printf("Total commits: %d (all branches)\n", len(commits))
	} else {
		fmt.Printf("Total commits: %d (branch: %s)\n", len(commits), currentBranch.Name)
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
