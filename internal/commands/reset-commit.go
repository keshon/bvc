package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// ResetCommitCommand reverts the workspace to a specific commit
type ResetCommitCommand struct{}

// Name returns the canonical Git-like command
func (c *ResetCommitCommand) Name() string { return "reset" }

// Usage string
func (c *ResetCommitCommand) Usage() string { return "reset <commit-id>" }

// Short description
func (c *ResetCommitCommand) Description() string {
	return "Revert the workspace to a specific commit"
}

// Detailed description
func (c *ResetCommitCommand) DetailedDescription() string {
	return "Revert changes to a specific commit by creating a new commit pointing to the same fileset.\n" +
		"Use 'bvc log' to find the commit ID you want to revert to."
}

// Optional aliases
func (c *ResetCommitCommand) Aliases() []string { return []string{"back"} }

// One-letter shortcut
func (c *ResetCommitCommand) Short() string { return "B" }

// Run executes the command
func (c *ResetCommitCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit ID required")
	}
	targetID := ctx.Args[0]
	return revertToCommit(targetID)
}

// revertToCommit performs the actual workspace revert
func revertToCommit(targetID string) error {
	// Load target commit
	commitPath := filepath.Join(config.CommitsDir, targetID+".json")
	var target core.Commit
	if err := util.ReadJSON(commitPath, &target); err != nil {
		return fmt.Errorf("unknown commit: %s", targetID)
	}

	// Load fileset of the target commit
	fsPath := filepath.Join(config.FilesetsDir, target.FilesetID+".json")
	var fs snapshot.Fileset
	if err := util.ReadJSON(fsPath, &fs); err != nil {
		return err
	}

	fmt.Printf("Reverting workspace to commit %s...\n", targetID)

	// Restore files from fileset
	if err := file.RestoreAll(fs.Files, fmt.Sprintf("for commit %s", targetID)); err != nil {
		return err
	}

	// Create a new commit pointing to the same fileset
	branch, _ := core.CurrentBranch()
	parent, _ := core.LastCommitID(branch.Name)
	newCommit := core.Commit{
		ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
		Parents:   []string{parent},
		Branch:    branch.Name,
		Message:   fmt.Sprintf("Revert to %s", targetID),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: target.FilesetID,
	}

	// Save the new commit
	newCommitPath := filepath.Join(config.CommitsDir, newCommit.ID+".json")
	if err := util.WriteJSON(newCommitPath, newCommit); err != nil {
		return err
	}
	if err := core.SetLastCommit(branch.Name, newCommit.ID); err != nil {
		return err
	}

	fmt.Println("Created revert commit:", newCommit.ID)
	fmt.Println("Workspace reverted to commit", targetID)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&ResetCommitCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
