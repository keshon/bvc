package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"app/internal/cli"
	"app/internal/storage/file"
)

// AddCommand implements Git-like 'add'
type AddCommand struct{}

func (c *AddCommand) Name() string  { return "add" }
func (c *AddCommand) Usage() string { return "add <file|dir|.>" }
func (c *AddCommand) Description() string {
	return "Stage files or directories for the next commit"
}
func (c *AddCommand) DetailedDescription() string {
	return `Stage changes for commit.
Usage:
  add .          - stage new and modified files
  add -A or --all - stage all changes, including deletions
  add -u         - stage modifications and deletions (no new files)
  add <path>     - stage a specific file or directory`
}
func (c *AddCommand) Aliases() []string { return nil }
func (c *AddCommand) Short() string     { return "a" }

func (c *AddCommand) Run(ctx *cli.Context) error {
	includeAll := false // -A or --all
	updateOnly := false // -u

	for _, arg := range ctx.Args {
		if arg == "--all" || arg == "-A" {
			includeAll = true
		}
		if arg == "-u" {
			updateOnly = true
		}
	}

	args := filterNonFlags(ctx.Args)

	// if no paths provided, assume "."
	if len(args) == 0 {
		args = []string{"."}
	}

	var toStage []string
	for _, arg := range args {
		if arg == "." {
			all, err := file.ListAll()
			if err != nil {
				return err
			}
			toStage = append(toStage, all...)
		} else if strings.ContainsAny(arg, "*?") {
			// handle glob patterns like *.go
			matches, err := filepath.Glob(arg)
			if err != nil {
				return err
			}
			toStage = append(toStage, matches...)
		} else {
			toStage = append(toStage, arg)
		}
	}

	if len(toStage) == 0 {
		return fmt.Errorf("no matching files to add")
	}

	var entries []file.Entry
	if includeAll {
		entries, _ = file.BuildTrackedAndUntracked(toStage)
	} else if updateOnly {
		entries, _ = file.BuildModifiedAndDeleted(toStage)
	} else {
		entries, _ = file.BuildAll(toStage)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no changes to stage")
	}

	if err := file.StageFiles(entries); err != nil {
		return err
	}

	fmt.Printf("Staged %d file(s)\n", len(entries))
	return nil
}

// filterNonFlags removes CLI flags like -A or --all from args
func filterNonFlags(args []string) []string {
	var res []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		res = append(res, a)
	}
	return res
}

func init() {
	cli.RegisterCommand(&AddCommand{})
}
