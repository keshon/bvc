package cherry_pick

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
	"time"
)

type Command struct{}

func (c *Command) Name() string      { return "cherry-pick" }
func (c *Command) Short() string     { return "C" }
func (c *Command) Aliases() []string { return []string{"cp"} }
func (c *Command) Usage() string     { return "cherry-pick <commit-id>" }
func (c *Command) Brief() string     { return "Apply selected commit to the current branch" }
func (c *Command) Help() string {
	return `Apply a specific commit to the current branch.
Use 'bvc log all' to find the commit ID you want to apply.`
}

func (c *Command) Run(ctx *command.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit ID required")
	}
	commitID := ctx.Args[0]
	return pickCommit(commitID)
}

// pickCommit applies the target commit to the current branch
func pickCommit(commitID string) error {
	// Open the repository context
	r, err := repo.OpenAt(config.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	targetCommit, err := r.GetCommit(commitID)
	if err != nil {
		return err
	}

	targetFileset, err := r.Storage.Snapshots.Load(targetCommit.FilesetID)
	if err != nil {
		return err
	}

	// Get current branch
	GetCurrentBranch, err := r.GetCurrentBranch()
	if err != nil {
		return err
	}

	// Get parent commit
	parent, err := r.GetLastCommitID(GetCurrentBranch.Name)
	if err != nil {
		return err
	}

	// Create new commit on current branch referencing the picked commit
	newCommit := repo.Commit{
		ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
		Parents:   []string{parent},
		Branch:    GetCurrentBranch.Name,
		Message:   fmt.Sprintf("Pick commit %s", commitID),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: targetCommit.FilesetID,
	}

	// Create commit
	_, err = r.CreateCommit(&newCommit)
	if err != nil {
		return err
	}

	// Update last commit for the branch
	if err := r.SetLastCommitID(GetCurrentBranch.Name, newCommit.ID); err != nil {
		return err
	}

	// Restore files from picked commit
	if err := r.Storage.Files.Restore(targetFileset.Files, fmt.Sprintf("pick commit %s", commitID)); err != nil {
		return err
	}

	fmt.Printf("Picked commit %s into branch '%s' as %s\n", commitID, GetCurrentBranch.Name, newCommit.ID)
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
