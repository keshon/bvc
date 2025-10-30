package blocks

import (
	"fmt"
	"sort"
	"strings"

	"app/internal/cli"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/util"
)

type Command struct{}

func (c *Command) Name() string      { return "blocks" }
func (c *Command) Short() string     { return "B" }
func (c *Command) Aliases() []string { return []string{"block"} }
func (c *Command) Usage() string     { return "blocks [branch|name]" }
func (c *Command) Brief() string     { return "Display repository blocks overview" }
func (c *Command) Help() string {
	return `Show repository blocks list with optional sort:
  - default: by block hash
  - branch: sort by branch name
  - name: sort by file name

Useful for identifying shared blocks between branches and associated files.`
}

func (c *Command) Run(ctx *cli.Context) error {
	sortMode := "block"

	if len(ctx.Args) > 0 {
		sortMode = strings.ToLower(ctx.Args[0])
	}

	return blocksOverview(sortMode)
}

func blocksOverview(sortMode string) error {
	blocksMap, err := repo.ListAllBlocks(false)
	if err != nil {
		return err
	}

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
	cli.RegisterCommand(
		cli.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
