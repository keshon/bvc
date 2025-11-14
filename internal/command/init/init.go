package init

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/fs"

	"app/internal/middleware"
	"app/internal/repo"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Command struct{}

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
func (c *Command) Flags(fs *flag.FlagSet)         {}
func (c *Command) Subcommands() []command.Command { return nil }

func (c *Command) Run(ctx *command.Context) error {
	fset := flag.NewFlagSet("init", flag.ContinueOnError)

	quiet := fset.Bool("quiet", false, "")
	fset.BoolVar(quiet, "q", false, "alias for --quiet")
	sepDir := fset.String("separate-bvc-dir", "", "")
	initBranch := fset.String("initial-branch", config.DefaultBranch, "")
	fset.StringVar(initBranch, "b", config.DefaultBranch, "alias for --initial-branch")

	if err := fset.Parse(ctx.Args); err != nil {
		return err
	}

	// determine repoDir
	repoDir := config.ResolveRepoRoot()

	// initialize FS
	fs := fs.NewOSFS()

	// if --separate-bvc-dir is provided, override pointer
	if *sepDir != "" {
		repoDir = *sepDir
		linkFile := filepath.Join(".", config.RepoPointerFile)
		if err := fs.WriteFile(linkFile, []byte(*sepDir), 0o644); err != nil {
			return fmt.Errorf("failed to write separate-bvc-dir pointer file: %w", err)
		}
	}

	// check if repo already exists
	cfg := config.NewRepoConfig(repoDir)
	alreadyExists := repo.IsRepoExists(cfg.RepoRoot)

	// initialize repository
	r, err := repo.NewRepositoryByPath(repoDir)
	if err != nil {
		return fmt.Errorf("failed to init repository: %w", err)
	}

	// warn if initial branch was specified but repo already exists
	if alreadyExists && *initBranch != config.DefaultBranch {
		fmt.Fprintf(os.Stderr, "warning: re-init: ignored --initial-branch=%s\n", *initBranch)
	}

	// set HEAD only if new repo
	if !alreadyExists {
		if _, err := r.Meta.SetHeadRef(*initBranch); err != nil {
			return fmt.Errorf("failed to set initial branch %q: %w", *initBranch, err)
		}
	}

	// output messages
	if !*quiet {
		root := absPath(r.Config.RepoRoot)
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
