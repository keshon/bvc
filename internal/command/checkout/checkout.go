package checkout

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"flag"
	"fmt"
)

type Command struct{}

func (c *Command) Name() string  { return "checkout" }
func (c *Command) Brief() string { return "Switch to another branch" }
func (c *Command) Usage() string { return "checkout <branch-name>" }
func (c *Command) Help() string {
	return `Switch to another branch.

Usage:
  checkout <branch-name>`
}
func (c *Command) Aliases() []string              { return []string{"co"} }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet)         {}

func (c *Command) Run(ctx *command.Context) error {
	// require branch name
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	branchName := ctx.Args[0]

	// open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// ensure branch exists
	targetBranch, err := r.Meta.GetBranch(branchName)
	if err != nil {
		return err
	}

	// resolve its last commit
	commitID, err := r.Meta.GetLastCommitID(targetBranch.Name)
	if err != nil {
		return err
	}

	// case 1: handle empty branch
	if commitID == "" {
		if err := r.Store.Files.Restore(nil, fmt.Sprintf("empty branch '%s'", branchName)); err != nil {
			return err
		}
		if _, err := r.Meta.SetHeadRef(branchName); err != nil {
			return err
		}
		fmt.Println("Branch is empty, switched to", branchName)
		return nil
	}

	// case 2: handle non-empty branch
	// load commit and fileset
	commit, err := r.Meta.GetCommit(commitID)
	if err != nil {
		return fmt.Errorf("failed to load commit %s: %w", commitID, err)
	}

	fs, err := r.Store.Snapshots.Load(commit.FilesetID)
	if err != nil {
		return fmt.Errorf("failed to load fileset %s: %w", commit.FilesetID, err)
	}

	// restore files
	if err := r.Store.Files.Restore(fs.Files, fmt.Sprintf("branch '%s'", branchName)); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// update HEAD and last commit
	if _, err := r.Meta.SetHeadRef(branchName); err != nil {
		return err
	}
	if err := r.Meta.SetLastCommitID(branchName, commitID); err != nil {
		return err
	}

	fmt.Println("Switched to branch", branchName)
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
