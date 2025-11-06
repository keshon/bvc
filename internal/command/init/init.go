package init

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Command struct{}

func (c *Command) Name() string      { return "init" }
func (c *Command) Short() string     { return "i" }
func (c *Command) Aliases() []string { return []string{"initialize"} }
func (c *Command) Usage() string     { return "init [options]" }
func (c *Command) Brief() string     { return "Initialize a new repository" }
func (c *Command) Help() string {
	return `Initialize a new repository in the current directory.

Options:
  -q, --quiet                 Suppress normal output.
      --bare                  Create a bare repository.
      --object-format=<algo>  Hash algorithm: xxh3-128 or sha256 (default xxh3-128).
      --separate-bvc-dir=<d>  Store repository data in a separate directory.
  -b, --initial-branch=<name> Use a custom initial branch name (default: main).
  
Usage:
  bvc init [options]

Examples:
  bvc init
  bvc init -q
  bvc init --bare
  bvc init --separate-bvc-dir=~/.bvc
  bvc init --initial-branch=master
`
}

func (c *Command) Run(ctx *command.Context) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)

	quiet := fs.Bool("quiet", false, "")
	fs.BoolVar(quiet, "q", false, "alias for --quiet")
	bare := fs.Bool("bare", false, "")
	objectFmt := fs.String("object-format", config.DefaultHash, "")
	sepDir := fs.String("separate-bvc-dir", "", "")
	initBranch := fs.String("initial-branch", config.DefaultBranch, "")
	fs.StringVar(initBranch, "b", config.DefaultBranch, "alias for --initial-branch")

	if err := fs.Parse(ctx.Args); err != nil {
		return err
	}

	// Determine repoDir
	repoDir := config.RepoDir
	if *bare {
		// bare repository, all data in current dir (no working tree)
		repoDir = filepath.Join(".", config.RepoDir)
	} else if *sepDir != "" {
		// separate directory (like --separate-git-dir)
		repoDir = *sepDir
	}

	// If separate dir used and not bare, create pointer file in working directory
	if *sepDir != "" && !*bare {
		linkFile := filepath.Join(".", config.RepoPointerFile)
		if err := os.WriteFile(linkFile, []byte(*sepDir), 0o644); err != nil {
			return fmt.Errorf("failed to write separate-bvc-dir pointer file: %w", err)
		}
	}

	// Initialize repository
	r, created, err := repo.InitAt(repoDir, *objectFmt)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			if !*quiet {
				fmt.Printf("Reinitialized existing repository at %q\n", r.Root)
			}
			return nil
		}
		return err
	}

	// Set HEAD to initial branch
	if _, err := r.SetHeadRef(*initBranch); err != nil {
		return fmt.Errorf("failed to set initial branch %q: %w", *initBranch, err)
	}

	if !*quiet {
		if created {
			fmt.Printf("Initialized empty repository in %q\n", r.Root)
		} else {
			fmt.Printf("Reinitialized existing repository in %q\n", r.Root)
		}
	}

	return nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
