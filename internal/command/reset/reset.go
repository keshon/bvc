package reset

import (
	"flag"
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct {
	soft  bool
	mixed bool
	hard  bool
}

func (c *Command) Name() string      { return "reset" }
func (c *Command) Aliases() []string { return []string{"drop"} }
func (c *Command) Usage() string     { return "reset [<commit-id>] [--soft|--mixed|--hard]" }
func (c *Command) Brief() string     { return "Reset current branch to a commit or HEAD" }
func (c *Command) Help() string {
	return `Reset current branch.

Options:
  --soft  : move HEAD only
  --mixed : move HEAD and reset index (default)
  --hard  : move HEAD, reset index and working directory

If <commit-id> is omitted, the last commit is used.

Usage:
  bvc reset [<commit-id>] [--soft|--mixed|--hard]

Examples:
  bvc reset
  bvc reset --mixed
  bvc reset --hard

  bvc reset <commit-id>
  bvc reset --soft <commit-id>
  bvc reset --mixed <commit-id>
  bvc reset --hard <commit-id>
`
}

func (c *Command) Subcommands() []command.Command { return nil }

func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.soft, "soft", false, "move HEAD only")
	fs.BoolVar(&c.mixed, "mixed", false, "move HEAD and reset index (default)")
	fs.BoolVar(&c.hard, "hard", false, "move HEAD, reset index and working directory")
}

func (c *Command) Run(ctx *command.Context) error {
	soft := c.soft
	mixed := c.mixed
	hard := c.hard

	mode := "mixed" // default mode
	count := 0

	if soft {
		mode = "soft"
		count++
	}
	if mixed {
		mode = "mixed"
		count++
	}
	if hard {
		mode = "hard"
		count++
	}

	if count > 1 {
		return fmt.Errorf("conflicting reset modes specified")
	}

	// repo open once
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	// extract commit-id argument
	targetID := ""
	if len(ctx.Args) > 0 {
		targetID = ctx.Args[0]
	}

	// if no commit specified — use last
	if targetID == "" {
		last, err := r.Meta.GetLastCommitID(branch.Name)
		if err != nil {
			return fmt.Errorf("cannot determine last commit: %w", err)
		}
		if last == "" {
			return fmt.Errorf("no commits to reset to")
		}
		targetID = last
	}

	// validate commit exists
	target, err := r.Meta.GetCommit(targetID)
	if err != nil {
		return fmt.Errorf("unknown commit: %s", targetID)
	}

	return c.reset(r, branch.Name, targetID, target.FilesetID, mode)
}

func (c *Command) reset(r *repo.Repository, branchName, targetID, filesetID, mode string) error {
	fmt.Printf("Resetting branch '%s' to commit %s (%s)...\n", branchName, targetID, mode)

	// move HEAD for all modes
	if err := r.Meta.SetLastCommitID(branchName, targetID); err != nil {
		return err
	}

	switch mode {
	case "soft":
		// nothing else
	case "mixed":
		if err := c.resetIndex(r, filesetID); err != nil {
			return err
		}
	case "hard":
		if err := c.resetIndex(r, filesetID); err != nil {
			return err
		}
		if err := c.resetWorkingDirectory(r, filesetID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported reset mode: %s", mode)
	}

	fmt.Println("Reset complete.")
	return nil
}

func (c *Command) resetIndex(r *repo.Repository, filesetID string) error {
	fs, err := r.Store.SnapshotCtx.Load(filesetID)
	if err != nil {
		return err
	}

	if err := r.Store.FileCtx.ClearIndex(); err != nil {
		return err
	}
	if err := r.Store.FileCtx.SaveIndexReplace(fs.Files); err != nil {
		return err
	}

	fmt.Println("Index reset.")
	return nil
}

func (c *Command) resetWorkingDirectory(r *repo.Repository, filesetID string) error {
	fs, err := r.Store.SnapshotCtx.Load(filesetID)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("reset --hard to fileset %s", filesetID)

	if err := r.Store.FileCtx.RestoreFilesToWorkingTree(fs.Files, msg); err != nil {
		return err
	}

	fmt.Println("Working directory reset.")
	return nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
			middleware.WithBlockIntegrityCheck(),
		),
	)
}
