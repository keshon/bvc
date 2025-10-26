package commands

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/storage"
	"app/internal/util"
)

type BlockCommand struct{}

func (c *BlockCommand) Name() string        { return "block" }
func (c *BlockCommand) Description() string { return "Display repository blocks overview" }
func (c *BlockCommand) Usage() string       { return "block [branch|name]" }
func (c *BlockCommand) DetailedDescription() string {
	return "Show repository blocks list with sort by block (default), branch, or file name.\nUseful if you need to identify which blocks are shared between branches and what files they cover."
}
func (c *BlockCommand) Run(ctx *cli.Context) error {
	sortMode := "block"
	if len(ctx.Args) > 0 {
		sortMode = strings.ToLower(ctx.Args[0])
	}
	return overviewBlocks(sortMode)
}

func overviewBlocks(sortMode string) error {
	branches, err := core.Branches()
	if err != nil {
		return err
	}
	sort.Strings(branches)

	type BlockInfo struct {
		Branches map[string]struct{}
		Files    map[string]struct{}
		Size     int64
	}
	blocks := map[string]*BlockInfo{}

	for _, branch := range branches {
		commitID, _ := core.LastCommit(branch)
		if commitID == "" {
			continue
		}
		var commit core.Commit
		if err := util.ReadJSON(filepath.Join(config.CommitsDir, commitID+".json"), &commit); err != nil {
			continue
		}

		var fs storage.Fileset
		if err := util.ReadJSON(filepath.Join(config.FilesetsDir, commit.FilesetID+".json"), &fs); err != nil {
			continue
		}

		for _, f := range fs.Files {
			for _, b := range f.Blocks {
				info, ok := blocks[b.Hash]
				if !ok {
					info = &BlockInfo{
						Branches: map[string]struct{}{},
						Files:    map[string]struct{}{},
						Size:     b.Size,
					}
					blocks[b.Hash] = info
				}
				info.Branches[branch] = struct{}{}
				info.Files[filepath.Base(f.Path)] = struct{}{}
			}
		}
	}

	type Row struct {
		Block    string
		Files    []string
		Branches []string
	}

	var rows []Row
	for hash, info := range blocks {
		fileList := util.SortedKeys(info.Files)
		branchList := util.SortedKeys(info.Branches)
		rows = append(rows, Row{
			Block:    hash,
			Files:    fileList,
			Branches: branchList,
		})
	}

	// Sorting modes
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
	default: // "block"
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Block < rows[j].Block
		})
	}

	fmt.Printf("Blocks overview (sorted by %s)\n", sortMode)
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))
	fmt.Printf("\033[90m%-32s %-32s %-32s\033[0m\n", "Block", "Name", "Branch")
	fmt.Println(strings.Repeat("\033[90m─\033[0m", 72))

	for _, r := range rows {
		name := truncateMid(strings.Join(r.Files, ","), 32)
		branch := truncateMid(strings.Join(r.Branches, ","), 32)
		fmt.Printf("\033[90m%-32s\033[0m %-32s %-32s\n", r.Block, name, branch)
	}

	// Orphan check
	objFiles, _ := filepath.Glob(filepath.Join(config.ObjectsDir, "*.bin"))
	var orphaned []string
	for _, f := range objFiles {
		h := filepath.Base(f)
		h = strings.TrimSuffix(h, ".bin")
		if _, ok := blocks[h]; !ok {
			orphaned = append(orphaned, h)
		}
	}
	if len(orphaned) > 0 {
		sort.Strings(orphaned)
		fmt.Println("\nOrphaned blocks:")
		for _, h := range orphaned {
			fmt.Printf("\033[90m%s\033[0m\n", h)
		}
	}

	return nil
}

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

func init() {
	cli.RegisterCommand(&BlockCommand{})
}
