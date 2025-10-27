package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
)

type LogCommand struct{}

// Canonical name
func (c *LogCommand) Name() string { return "log" }

// Usage string
func (c *LogCommand) Usage() string {
	return "log [-a|--all]"
}

// Short description
func (c *LogCommand) Description() string {
	return "Show commit history (current branch by default)"
}

// Detailed description
func (c *LogCommand) DetailedDescription() string {
	return "List commits for the current branch or all branches if -a / --all is specified."
}

// Optional aliases
func (c *LogCommand) Aliases() []string { return []string{"lg"} }

// One-letter shortcut
func (c *LogCommand) Short() string { return "l" }

// Run executes the log command
func (c *LogCommand) Run(ctx *cli.Context) error {
	showAll := false
	if _, ok := ctx.Flags["a"]; ok {
		showAll = true
	} else if _, ok := ctx.Flags["all"]; ok {
		showAll = true
	} else if len(ctx.Args) > 0 && ctx.Args[0] == "all" {
		// Positional fallback for backward compatibility
		showAll = true
	}

	return c.listCommits(showAll)
}

// LogRow holds structured commit information for printing
type LogRow struct {
	ID        string
	Date      string
	Branch    string
	Parent    string
	Message   string
	Timestamp time.Time
}

// listCommits gathers and prints commits in descending order
func (c *LogCommand) listCommits(showAll bool) error {
	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	var branchNames []string
	if showAll {
		entries, err := os.ReadDir(config.BranchesDir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			branchNames = append(branchNames, e.Name())
		}
	} else {
		branchNames = []string{currentBranch.Name}
	}

	var commits []*core.Commit
	seen := make(map[string]bool)

	for _, branch := range branchNames {
		_ = core.GetBranchCommits(branch, func(cmt *core.Commit) bool {
			if seen[cmt.ID] {
				return true
			}
			seen[cmt.ID] = true
			commits = append(commits, cmt)
			return true
		})
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

		// Print message with Git-style indentation
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

// Register the command
func init() {
	cli.RegisterCommand(&LogCommand{})
}
