package checkout

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
)

type Command struct{}

func (c *Command) Name() string      { return "checkout" }
func (c *Command) Short() string     { return "C" }
func (c *Command) Aliases() []string { return []string{"co"} }
func (c *Command) Usage() string     { return "checkout <branch-name>" }
func (c *Command) Brief() string     { return "Switch to another branch" }
func (c *Command) Help() string {
	return `Switch to another branch.

Usage:
  checkout <branch-name>`
}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	branchName := ctx.Args[0]

	// open the repository context
	r, err := repo.OpenAt(config.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// ensure branch exists
	targetBranch, err := r.GetBranch(branchName)
	if err != nil {
		return err
	}

	// resolve its last commit
	commitID, err := r.GetLastCommitID(targetBranch.Name)
	if err != nil {
		return err
	}

	// case 1: handle empty branch
	if commitID == "" {
		if err := r.Storage.Files.Restore(nil, fmt.Sprintf("empty branch '%s'", branchName)); err != nil {
			return err
		}
		if _, err := r.SetHeadRef(branchName); err != nil {
			return err
		}
		fmt.Println("Branch is empty, switched to", branchName)
		return nil
	}

	// case 2: handle non-empty branch
	// load commit and fileset
	commit, err := r.GetCommit(commitID)
	if err != nil {
		return fmt.Errorf("failed to load commit %s: %w", commitID, err)
	}

	fs, err := r.Storage.Snapshots.Load(commit.FilesetID)

	if err != nil {
		return fmt.Errorf("failed to load fileset %s: %w", commit.FilesetID, err)
	}

	// restore files
	if err := r.Storage.Files.Restore(fs.Files, fmt.Sprintf("branch '%s'", branchName)); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// update HEAD and last commit
	if _, err := r.SetHeadRef(branchName); err != nil {
		return err
	}
	if err := r.SetLastCommitID(branchName, commitID); err != nil {
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
