package commands

import (
	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type LogCommand struct{}

func (c *LogCommand) Name() string  { return "log" }
func (c *LogCommand) Usage() string { return "log [all]" }
func (c *LogCommand) Description() string {
	return "Show commit history (use 'all' for all branches)"
}
func (c *LogCommand) DetailedDescription() string {
	return "List commits for the current branch or all branches if 'all' is specified."
}

func (c *LogCommand) Run(ctx *cli.Context) error {
	showAll := len(ctx.Args) > 0 && ctx.Args[0] == "all"
	return listCommits(showAll)
}

type LogRow struct {
	ID        string
	Date      string
	Branch    string
	Parent    string
	Message   string
	Timestamp time.Time
}

func listCommits(showAll bool) error {
	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	// Determine which branches to list
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

	// Collect commits from all relevant branches
	for _, branch := range branchNames {
		_ = core.GetBranchCommits(branch, func(c *core.Commit) bool {
			if seen[c.ID] {
				return true
			}
			seen[c.ID] = true

			parent := "<none>"
			if len(c.Parents) > 0 {
				parent = strings.Join(c.Parents, ", ")
			}

			t, _ := time.Parse(time.RFC3339, c.Timestamp)
			rows = append(rows, LogRow{
				ID:        c.ID,
				Date:      c.Timestamp,
				Branch:    c.Branch,
				Parent:    parent,
				Message:   c.Message,
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

func init() {
	cli.RegisterCommand(&LogCommand{})
}
