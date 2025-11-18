package add

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/internal/repo/store/file"
)

type Command struct {
	all    bool
	update bool
}

func (c *Command) Name() string      { return "add" }
func (c *Command) Aliases() []string { return nil }
func (c *Command) Brief() string     { return "Stage files or directories for the next commit" }
func (c *Command) Usage() string     { return "add <file|dir|.> [options]" }
func (c *Command) Help() string {
	return `Stage changes for commit.

Options:
  -a, --all             Stage all changes, including deletions (-A)
	  --update          Stage modifications and deletions only (-u)

Usage:
  bvc add <file|dir|.> [options]

Examples:
  bvc add .
  bvc add 'main.go'
  bvc add dir/
`
}
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.all, "all", false, "Stage all changes, including deletions (-A)")
	fs.BoolVar(&c.update, "update", false, "Stage modifications and deletions only (-u)")
}

func (c *Command) Run(ctx *command.Context) error {
	includeAll := c.all
	updateOnly := c.update

	args := filterNonFlags(ctx.Args)
	if len(args) == 0 {
		args = []string{"."}
	}

	// Open repository
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Collect repo filesets (working, staged, ignored)
	trackedFS, stagedFS, _, err := r.Store.SnapshotCtx.BuildAllRepositoryFilesets()
	if err != nil {
		return fmt.Errorf("failed to scan repository files: %w", err)
	}

	var entries []file.Entry

	switch {
	case includeAll:
		// Stage all tracked changes (new, modified, deleted)
		entries = trackedFS.Files

	case updateOnly:
		// Stage only modifications and deletions for already-staged files
		stagedMap := make(map[string]file.Entry, len(stagedFS.Files))
		for _, e := range stagedFS.Files {
			stagedMap[e.Path] = e
		}
		for _, e := range trackedFS.Files {
			if _, exists := stagedMap[e.Path]; exists {
				entries = append(entries, e)
			}
		}

	default:
		// Stage specific paths or globs
		for _, arg := range args {
			matches := filterMatchingEntries(trackedFS.Files, arg)
			entries = append(entries, matches...)
		}
	}

	if len(entries) == 0 {
		return fmt.Errorf("no changes to stage")
	}

	// Write staged entries to index
	if err := r.Store.FileCtx.SaveIndexMerge(entries); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
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

func filterMatchingEntries(entries []file.Entry, pattern string) []file.Entry {
	var matched []file.Entry

	if pattern == "." {
		return entries
	}

	for _, e := range entries {
		ok, _ := filepath.Match(pattern, filepath.Base(e.Path))
		if ok || strings.HasPrefix(e.Path, pattern) {
			matched = append(matched, e)
		}
	}
	return matched
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
