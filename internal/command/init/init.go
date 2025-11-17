package init

import (
	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"

	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct {
	quiet          bool
	separateBvcDir string
	initialBranch  string
}

func (c *Command) Name() string      { return "init" }
func (c *Command) Short() string     { return "i" }
func (c *Command) Aliases() []string { return []string{} }
func (c *Command) Usage() string     { return "init [options]" }
func (c *Command) Brief() string     { return "Initialize a new repository" }
func (c *Command) Help() string {
	return `Initialize a new repository in the current directory.

Options:
  -q, --quiet                 Suppress normal output.
      --separate-bvc-dir=<d>  Store repository data in a separate directory.
  -b, --initial-branch=<name> Use a custom initial branch name (default: main).
  
Usage:
  bvc init [options]

Examples:
  bvc init
  bvc init -q
  bvc init --separate-bvc-dir=~/.bvc
  bvc init --initial-branch=master
`
}
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.quiet, "quiet", false, "Suppress normal output.")

	fs.StringVar(&c.separateBvcDir, "separate-bvc-dir", "", "Store repository data in a separate directory.")

	fs.StringVar(&c.initialBranch, "initial-branch", config.DefaultBranch, "Use a custom initial branch name (default: main).")
}
func (c *Command) Subcommands() []command.Command { return nil }

func (c *Command) Run(ctx *command.Context) error {
	quiet := c.quiet
	sepDir := c.separateBvcDir
	initBranch := c.initialBranch

	repoDir := config.ResolveRepoDir()
	fs := fs.NewOSFS()

	// if --separate-bvc-dir is provided, override pointer
	if sepDir != "" {
		repoDir = sepDir
		linkFile := filepath.Join(".", config.RepoPointerFile)
		if err := fs.WriteFile(linkFile, []byte(sepDir), 0o644); err != nil {
			return fmt.Errorf("failed to write separate-bvc-dir pointer file: %w", err)
		}
	}

	// check if repo already exists
	cfg := config.NewRepoConfig(repoDir)
	alreadyExists := repo.IsRepoExists(cfg.RepoDir)

	// initialize repository
	r, err := repo.NewRepositoryByPath(repoDir)
	if err != nil {
		return fmt.Errorf("failed to init repository: %w", err)
	}

	// warn if initial branch was specified but repo already exists
	if alreadyExists && initBranch != config.DefaultBranch {
		fmt.Fprintf(os.Stderr, "warning: re-init: ignored --initial-branch=%s\n", initBranch)
	}

	// set HEAD only if new repo
	if !alreadyExists {
		if _, err := r.Meta.SetHeadRef(initBranch); err != nil {
			return fmt.Errorf("failed to set initial branch %q: %w", initBranch, err)
		}
	}

	// output messages
	if !quiet {
		root := absPath(r.Config.RepoDir)
		if alreadyExists {
			fmt.Printf("Reinitialized existing BVC repository in %s\n", root)
		} else {
			fmt.Printf("Initialized empty BVC repository in %s\n", root)
		}
	}

	return nil
}

// helper to convert paths to absolute form
func absPath(path string) string {
	if p, err := filepath.Abs(path); err == nil {
		return p
	}
	return path
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
