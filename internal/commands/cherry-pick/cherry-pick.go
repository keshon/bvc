package cherry_pick

import (
	"fmt"
	"time"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
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

func (c *Command) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit ID required")
	}
	commitID := ctx.Args[0]
	return pickCommit(commitID)
}

// pickCommit applies the target commit to the current branch
func pickCommit(commitID string) error {
	targetCommit, err := core.GetCommit(commitID)
	if err != nil {
		return err
	}
	targetFileset, err := snapshot.GetFileset(targetCommit.FilesetID)
	if err != nil {
		return err
	}

	// Get current branch
	GetCurrentBranch, err := core.GetCurrentBranch()
	if err != nil {
		return err
	}

	// Get parent commit
	parent, err := core.GetLastCommitID(GetCurrentBranch.Name)
	if err != nil {
		return err
	}

	// Create new commit on current branch referencing the picked commit
	newCommit := core.Commit{
		ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
		Parents:   []string{parent},
		Branch:    GetCurrentBranch.Name,
		Message:   fmt.Sprintf("Pick commit %s", commitID),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: targetCommit.FilesetID,
	}

	// Create commit
	_, err = core.CreateCommit(&newCommit)
	if err != nil {
		return err
	}

	// Update last commit for the branch
	if err := core.SetLastCommitID(GetCurrentBranch.Name, newCommit.ID); err != nil {
		return err
	}

	// Restore files from picked commit
	if err := file.RestoreFiles(targetFileset.Files, fmt.Sprintf("pick commit %s", commitID)); err != nil {
		return err
	}

	fmt.Printf("Picked commit %s into branch '%s' as %s\n", commitID, GetCurrentBranch.Name, newCommit.ID)
	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
			middleware.WithBlockIntegrityCheck(),
		),
	)
}
