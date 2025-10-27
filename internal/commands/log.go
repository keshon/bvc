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

	// Determine which branches to process
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

	var rows []LogRow
	seen := make(map[string]bool)

	// Collect commits from each branch
	for _, branch := range branchNames {
		_ = core.GetBranchCommits(branch, func(cmt *core.Commit) bool {
			if seen[cmt.ID] {
				return true
			}
			seen[cmt.ID] = true

			parent := "<none>"
			if len(cmt.Parents) > 0 {
				parent = strings.Join(cmt.Parents, ", ")
			}

			t, _ := time.Parse(time.RFC3339, cmt.Timestamp)
			rows = append(rows, LogRow{
				ID:        cmt.ID,
				Date:      cmt.Timestamp,
				Branch:    cmt.Branch,
				Parent:    parent,
				Message:   cmt.Message,
				Timestamp: t,
			})
			return true
		})
	}

	if len(rows) == 0 {
		fmt.Println("No commits found")
		return nil
	}

	// Sort newest first
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Timestamp.After(rows[j].Timestamp)
	})

	// Output formatting
	fmt.Println("Commits history")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 100))
	fmt.Printf("\033[90m%-16s  %-19s  %-8s  %-28s  %s\033[0m\n", "ID", "Date", "Branch", "Parent(s)", "Message")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 100))

	for _, r := range rows {
		parent := r.Parent
		if len(parent) > 28 {
			parent = parent[:25] + "..."
		}
		fmt.Printf("\033[90m%-16s\033[0m  %-19s  %-8s  %-28s  %s\n",
			r.ID, r.Timestamp.Format("2006-01-02 15:04:05"), r.Branch, parent, r.Message)
	}

	if showAll {
		fmt.Printf("\nTotal commits: %d (all branches)\n", len(rows))
	} else {
		fmt.Printf("\nTotal commits: %d (branch: %s)\n", len(rows), currentBranch.Name)
	}

	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(&LogCommand{})
}
