package cherry_pick

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

// Command applies a specific commit to the current branch
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "cherry-pick" }

// Usage string
func (c *Command) Usage() string { return "cherry-pick <commit-id>" }

// Short description
func (c *Command) Description() string {
	return "Apply selected commit to the current branch"
}

// Detailed description
func (c *Command) DetailedDescription() string {
	return `Apply a specific commit to the current branch.
Use 'bvc log all' to find the commit ID you want to apply.`
}

// Optional aliases
func (c *Command) Aliases() []string { return []string{"cp"} }

// One-letter shortcut
func (c *Command) Short() string { return "C" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit ID required")
	}
	commitID := ctx.Args[0]
	return pickCommit(commitID)
}

// pickCommit applies the target commit to the current branch
func pickCommit(commitID string) error {
	// Load target commit
	targetPath := filepath.Join(config.CommitsDir, commitID+".json")
	var target core.Commit
	if err := util.ReadJSON(targetPath, &target); err != nil {
		return fmt.Errorf("unknown commit: %s", commitID)
	}

	// Load fileset of target commit
	fsPath := filepath.Join(config.FilesetsDir, target.FilesetID+".json")
	var fs snapshot.Fileset
	if err := util.ReadJSON(fsPath, &fs); err != nil {
		return err
	}

	// Get current branch
	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	// Get parent commit
	parent, err := core.LastCommitID(currentBranch.Name)
	if err != nil {
		return err
	}

	// Create new commit on current branch referencing the picked commit
	newCommit := core.Commit{
		ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
		Parents:   []string{parent},
		Branch:    currentBranch.Name,
		Message:   fmt.Sprintf("Pick commit %s", commitID),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: target.FilesetID,
	}

	newCommitPath := filepath.Join(config.CommitsDir, newCommit.ID+".json")
	if err := util.WriteJSON(newCommitPath, newCommit); err != nil {
		return err
	}

	// Update last commit for the branch
	if err := core.SetLastCommit(currentBranch.Name, newCommit.ID); err != nil {
		return err
	}

	// Restore files from picked commit
	if err := file.RestoreFiles(fs.Files, fmt.Sprintf("pick commit %s", commitID)); err != nil {
		return err
	}

	fmt.Printf("Picked commit %s into branch '%s' as %s\n", commitID, currentBranch.Name, newCommit.ID)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&Command{}, middleware.WithBlockIntegrityCheck()),
	)
}
