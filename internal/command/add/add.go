package add

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repo/store/file"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

type Command struct{}

func (c *Command) Name() string  { return "add" }
func (c *Command) Brief() string { return "Stage files or directories for the next commit" }
func (c *Command) Usage() string { return "add <file|dir|.> [options]" }
func (c *Command) Help() string {
	return `Stage changes for commit.

Usage:
  add .              - stage new and modified files
  add -A or --all    - stage all changes, including deletions
  add -u or --update - stage modifications and deletions (no new files)
  add <path>         - stage a specific file or directory`
}
func (c *Command) Aliases() []string              { return nil }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.Bool("all", false, "Stage all changes, including deletions (-A)")
	fs.Bool("update", false, "Stage modifications and deletions only (-u)")
}

func (c *Command) Run(ctx *command.Context) error {
	includeAll := ctx.Flags.Lookup("all").Value.(flag.Getter).Get().(bool)
	updateOnly := ctx.Flags.Lookup("update").Value.(flag.Getter).Get().(bool)

	args := filterNonFlags(ctx.Args)
	if len(args) == 0 {
		args = []string{"."}
	}

	// open repository
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	var toStage []string
	for _, arg := range args {
		if arg == "." {
			paths, _, err := r.Store.Files.ScanFilesInWorkingTree()
			if err != nil {
				return err
			}
			toStage = append(toStage, paths...)
		} else if strings.ContainsAny(arg, "*?") {
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
		entries, err = r.Store.Files.BuildAllEntries()
	} else if updateOnly {
		entries, err = r.Store.Files.BuildChangedEntries()
	} else {
		entries, err = r.Store.Files.BuildEntries(toStage)
	}
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("no changes to stage")
	}

	if err := r.Store.Files.SaveIndex(entries); err != nil {
		return err
	}

	fmt.Printf("Staged %d file(s)\n", len(entries))
	return nil
}

// helper: remove flags from args
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
			middleware.WithBlockIntegrityCheck(),
			middleware.WithDebugArgsPrint(),
		),
	)
}
