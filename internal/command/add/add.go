package add

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repo/store/file"
	"fmt"
	"path/filepath"
	"strings"
)

type Command struct{}

func (c *Command) Name() string      { return "add" }
func (c *Command) Short() string     { return "a" }
func (c *Command) Aliases() []string { return nil }
func (c *Command) Usage() string     { return "add <file|dir|.>" }
func (c *Command) Brief() string     { return "Stage files or directories for the next commit" }
func (c *Command) Help() string {
	return `Stage changes for commit.

Usage:
  add .              - stage new and modified files
  add -A or --all    - stage all changes, including deletions
  add -u or --update - stage modifications and deletions (no new files)
  add <path>         - stage a specific file or directory`
}

func (c *Command) Run(ctx *command.Context) error {
	includeAll := false // -A or --all
	updateOnly := false // -u or --update

	for _, arg := range ctx.Args {
		if arg == "--all" || arg == "-A" {
			includeAll = true
		}
		if arg == "--update" || arg == "-u" {
			updateOnly = true
		}
	}

	args := filterNonFlags(ctx.Args)

	// if no paths provided, assume "."
	if len(args) == 0 {
		args = []string{"."}
	}

	// open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	var toStage []string
	for _, arg := range args {
		if arg == "." {
			all, err := r.Store.Files.ListAll()
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

	// create staged entries
	var entries []file.Entry
	if includeAll {
		entries, err = r.Store.Files.CreateAllEntries()
	} else if updateOnly {
		entries, err = r.Store.Files.CreateChangedEntries()
	} else {
		entries, err = r.Store.Files.CreateEntries(toStage)
	}
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("no changes to stage")
	}

	if err := r.Store.Files.StageFiles(entries); err != nil {
		return err
	}

	fmt.Printf("Staged %d file(s)\n", len(entries))
	return nil
}

// filterNonFlags removes CLI flags starting with "-" or "--"
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
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
