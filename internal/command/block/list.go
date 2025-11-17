package block

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/internal/repotools"
	"github.com/keshon/bvc/internal/util"
)

type ListCommand struct{}

func (c *ListCommand) Name() string                   { return "list" }
func (c *ListCommand) Brief() string                  { return "Display repository blocks list" }
func (c *ListCommand) Usage() string                  { return "block list [branch|name]" }
func (c *ListCommand) Help() string                   { return "Show repository blocks list" }
func (c *ListCommand) Aliases() []string              { return []string{"bl"} }
func (c *ListCommand) Subcommands() []command.Command { return nil }
func (c *ListCommand) Flags(fs *flag.FlagSet)         {}

func (c *ListCommand) Run(ctx *command.Context) error {
	sortMode := "block"
	if len(ctx.Args) > 0 {
		sortMode = strings.ToLower(ctx.Args[0])
	}

	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	blocksMap, err := repotools.ListAllBlocks(r.Meta, r.Config, true)
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
		sort.Slice(rows, func(i, j int) bool { return rows[i].Hash < rows[j].Hash })
	}

	fmt.Printf("Blocks list (sorted by %s)\n", sortMode)
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 110))
	fmt.Printf("\033[90m%-32s %-32s %-32s\033[0m\n", "Block", "Name", "Branch")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 110))

	for _, row := range rows {
		var files []string
		for _, f := range row.Files {
			files = append(files, strings.TrimPrefix(f, r.Config.WorkingTreeDir))
		}
		row.Files = files

		name := truncateStringInMid(strings.Join(row.Files, ","), 70)
		branch := truncateStringInMid(strings.Join(row.Branches, ","), 70)
		fmt.Printf("\033[90m%-32s\033[0m %-32s \033[90m%-32s\033[0m\n", row.Hash, name, branch)
	}

	return nil
}

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
			&BlockCommand{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
