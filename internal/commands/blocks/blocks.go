package blocks

import (
	"fmt"
	"sort"
	"strings"

	"app/internal/cli"
	"app/internal/repo"
	"app/internal/util"
)

// Command displays repository block overview
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "blocks" }

// Usage string
func (c *Command) Usage() string {
	return "blocks [branch|name]"
}

// Short description
func (c *Command) Description() string {
	return "Display repository blocks overview"
}

// Detailed description
func (c *Command) DetailedDescription() string {
	return `Show repository blocks list with optional sort:
  - default: by block hash
  - branch: sort by branch name
  - name: sort by file name

Useful for identifying shared blocks between branches and associated files.`
}

// Optional aliases
func (c *Command) Aliases() []string { return []string{"block"} }

// One-letter shortcut
func (c *Command) Short() string { return "B" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	sortMode := "block"

	if len(ctx.Args) > 0 {
		sortMode = strings.ToLower(ctx.Args[0])
	}

	return c.overviewBlocks(sortMode)
}

// overviewBlocks collects blocks and prints the table
func (c *Command) overviewBlocks(sortMode string) error {
	// Collect all blocks from repo
	blocksMap, err := repo.ListAllBlocks(false)
	if err != nil {
		return err
	}

	// Prepare rows
	type Row struct {
		Hash     string
		Files    []string
		Branches []string
	}

	rows := make([]Row, 0, len(blocksMap))
	for hash, info := range blocksMap {
		rows = append(rows, Row{
			Hash:     hash,
			Files:    util.SortedKeys(info.Files),
			Branches: util.SortedKeys(info.Branches),
		})
	}

	// Sorting logic
	switch sortMode {
	case "branch":
		sort.Slice(rows, func(i, j int) bool {
			if len(rows[i].Branches) == 0 {
				return false
			}
			if len(rows[j].Branches) == 0 {
				return true
			}
			return rows[i].Branches[0] < rows[j].Branches[0]
		})
	case "name":
		sort.Slice(rows, func(i, j int) bool {
			if len(rows[i].Files) == 0 {
				return false
			}
			if len(rows[j].Files) == 0 {
				return true
			}
			return rows[i].Files[0] < rows[j].Files[0]
		})
	default:
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Hash < rows[j].Hash
		})
	}

	// Print table header
	fmt.Printf("Blocks overview (sorted by %s)\n", sortMode)
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))
	fmt.Printf("\033[90m%-32s %-32s %-32s\033[0m\n", "Block", "Name", "Branch")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))

	// Print rows
	for _, r := range rows {
		name := truncateMid(strings.Join(r.Files, ","), 32)
		branch := truncateMid(strings.Join(r.Branches, ","), 32)
		fmt.Printf("\033[90m%-32s\033[0m %-32s %-32s\n", r.Hash, name, branch)
	}

	return nil
}

// truncateMid shortens long strings with "..." in the middle
func truncateMid(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 6 {
		return s[:width]
	}
	half := (width - 3) / 2
	return s[:half] + "..." + s[len(s)-half:]
}

// Register the command
func init() {
	cli.RegisterCommand(&Command{})
}
