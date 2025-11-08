package blocks

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repotools"
	"app/internal/util"
	"fmt"
	"sort"
	"strings"
)

type Command struct{}

func (c *Command) Name() string      { return "blocks" }
func (c *Command) Short() string     { return "B" }
func (c *Command) Aliases() []string { return []string{"block"} }
func (c *Command) Usage() string     { return "blocks [branch|name]" }
func (c *Command) Brief() string     { return "Display repository blocks overview" }
func (c *Command) Help() string {
	return `Show repository blocks list with optional sort mode.

Usage:
  blocks        - show all blocks
  blocks branch - sort by branch name
  blocks name 	- sort by file name

Useful for identifying shared blocks between branches and associated files.`
}

func (c *Command) Run(ctx *command.Context) error {
	sortMode := "block" // default

	if len(ctx.Args) > 0 {
		sortMode = strings.ToLower(ctx.Args[0])
	}

	// open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	// list all blocks
	blocksMap, err := repotools.ListAllBlocks(r.Meta, r.Config, true)
	if err != nil {
		return err
	}

	// struct represents a row in output list
	type Row struct {
		Hash     string
		Files    []string
		Branches []string
	}

	// prepare rows
	rows := make([]Row, 0, len(blocksMap))
	for hash, info := range blocksMap {
		rows = append(rows, Row{
			Hash:     hash,
			Files:    util.SortedKeys(info.Files),
			Branches: util.SortedKeys(info.Branches),
		})
	}

	// sort rows
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

	// print rows
	fmt.Printf("Blocks overview (sorted by %s)\n", sortMode)
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))
	fmt.Printf("\033[90m%-32s %-32s %-32s\033[0m\n", "Block", "Name", "Branch")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))

	for _, r := range rows {
		name := truncateStringInMid(strings.Join(r.Files, ","), 32)
		branch := truncateStringInMid(strings.Join(r.Branches, ","), 32)
		fmt.Printf("\033[90m%-32s\033[0m %-32s %-32s\n", r.Hash, name, branch)
	}

	return nil
}

// truncateStringInMid shortens long strings with "..." in the middle
func truncateStringInMid(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 6 {
		return s[:width]
	}
	half := (width - 3) / 2
	return s[:half] + "..." + s[len(s)-half:]
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
