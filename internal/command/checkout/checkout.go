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
Restores the branch's fileset and updates HEAD reference.`
}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	branch := ctx.Args[0]
	return checkoutBranch(branch)
}

func checkoutBranch(branch string) error {
	// Open the repository context
	r, err := repo.OpenAt(config.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Ensure branch exists
	b, err := r.GetBranch(branch)
	if err != nil {
		return err
	}

	// Resolve its last commit
	commitID, err := r.GetLastCommitID(b.Name)
	if err != nil {
		return err
	}

	// Handle empty branch
	if commitID == "" {
		if err := r.Storage.Files.Restore(nil, fmt.Sprintf("empty branch '%s'", branch)); err != nil {
			return err
		}
		if _, err := r.SetHeadRef("branches/" + branch); err != nil {
			return err
		}
		fmt.Println("Branch is empty, switched to", branch)
		return nil
	}

	// Load commit and fileset
	commit, err := r.GetCommit(commitID)
	if err != nil {
		return fmt.Errorf("failed to load commit %s: %w", commitID, err)
	}

	fs, err := r.Storage.Snapshots.Load(commit.FilesetID)

	if err != nil {
		return fmt.Errorf("failed to load fileset %s: %w", commit.FilesetID, err)
	}

	// Restore files
	if err := r.Storage.Files.Restore(fs.Files, fmt.Sprintf("branch '%s'", branch)); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// Update HEAD and last commit
	if _, err := r.SetHeadRef("branches/" + branch); err != nil {
		return err
	}
	if err := r.SetLastCommitID(branch, commitID); err != nil {
		return err
	}

	fmt.Println("Switched to branch", branch)
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
