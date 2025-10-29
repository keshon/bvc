package checkout

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
)

// Command switches to another branch
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "checkout" }

// Usage string
func (c *Command) Usage() string { return "checkout <branch-name>" }

// Short description
func (c *Command) Brief() string { return "Switch to another branch" }

// Detailed description
func (c *Command) Help() string {
	return `Switch to another branch.
Restores the branch's fileset and updates HEAD reference.`
}

// Optional aliases
func (c *Command) Aliases() []string { return []string{"co"} }

// One-letter shortcut
func (c *Command) Short() string { return "C" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	branch := ctx.Args[0]
	return runCheckout(branch)
}

// runCheckout performs the actual branch switch using core and storage layers
func runCheckout(branch string) error {
	// Step 1: Ensure branch exists
	b, err := core.GetBranch(branch)
	if err != nil {
		return err
	}

	// Step 2: Resolve its last commit
	commitID, err := core.LastCommitID(b.Name)
	if err != nil {
		return err
	}

	// Step 3: Handle empty branch
	if commitID == "" {
		if err := file.RestoreFiles(nil, fmt.Sprintf("empty branch '%s'", branch)); err != nil {
			return err
		}
		if _, err := core.SetHeadRef("branches/" + branch); err != nil {
			return err
		}
		fmt.Println("Branch is empty, switched to", branch)
		return nil
	}

	// Step 4: Load commit and fileset
	commit, err := core.GetCommit(commitID)
	if err != nil {
		return fmt.Errorf("failed to load commit %s: %w", commitID, err)
	}

	fs, err := snapshot.LoadFileset(commit.FilesetID)
	if err != nil {
		return fmt.Errorf("failed to load fileset %s: %w", commit.FilesetID, err)
	}

	// Step 5: Restore files
	if err := file.RestoreFiles(fs.Files, fmt.Sprintf("branch '%s'", branch)); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// Step 6: Update HEAD and last commit
	if _, err := core.SetHeadRef("branches/" + branch); err != nil {
		return err
	}
	if err := core.SetLastCommit(branch, commitID); err != nil {
		return err
	}

	fmt.Println("Switched to branch", branch)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&Command{}, middleware.WithBlockIntegrityCheck()),
	)
}
