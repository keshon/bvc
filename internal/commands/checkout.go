package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/middleware"

	"app/internal/storage/file"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// CheckoutCommand switches to another branch
type CheckoutCommand struct{}

// Canonical name
func (c *CheckoutCommand) Name() string { return "checkout" }

// Usage string
func (c *CheckoutCommand) Usage() string { return "checkout <branch-name>" }

// Short description
func (c *CheckoutCommand) Description() string { return "Switch to another branch" }

// Detailed description
func (c *CheckoutCommand) DetailedDescription() string {
	return `Switch to another branch.
Restores the branch's fileset and updates HEAD reference.`
}

// Optional aliases
func (c *CheckoutCommand) Aliases() []string { return []string{"co"} }

// One-letter shortcut
func (c *CheckoutCommand) Short() string { return "C" }

// Run executes the command
func (c *CheckoutCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	branch := ctx.Args[0]
	return checkoutBranch(branch)
}

// checkoutBranch performs the actual branch switch
func checkoutBranch(branch string) error {
	branchPath := filepath.Join(config.BranchesDir, branch)

	// Check if branch exists
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		return fmt.Errorf("branch '%s' does not exist", branch)
	}

	commitIDBytes, err := os.ReadFile(branchPath)
	if err != nil {
		return err
	}
	commitID := string(commitIDBytes)

	// Empty branch handling
	if commitID == "" {
		emptyFS := snapshot.Fileset{ID: "empty", Files: nil}
		if err := file.RestoreAll(emptyFS.Files, fmt.Sprintf("empty branch '%s'", branch)); err != nil {
			return err
		}

		if _, err := core.SetHeadRef(filepath.Join("branches", branch)); err != nil {
			return err
		}

		fmt.Println("Branch is empty, switched to", branch)
		return nil
	}

	// Load commit fileset
	commitPath := filepath.Join(config.CommitsDir, commitID+".json")
	var commit core.Commit
	if err := util.ReadJSON(commitPath, &commit); err != nil {
		return err
	}

	fsPath := filepath.Join(config.FilesetsDir, commit.FilesetID+".json")
	var fileset snapshot.Fileset
	if err := util.ReadJSON(fsPath, &fileset); err != nil {
		return err
	}

	// Restore files
	if err := file.RestoreAll(fileset.Files, fmt.Sprintf("for branch '%s'", branch)); err != nil {
		return err
	}

	// Update HEAD and last commit
	if _, err := core.SetHeadRef(filepath.Join("branches", branch)); err != nil {
		return err
	}
	_ = core.SetLastCommit(branch, commitID)

	fmt.Println("Switched to branch", branch)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&CheckoutCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
