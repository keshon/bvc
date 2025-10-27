package commands

import (
	"fmt"
	"path/filepath"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// ResetHardCommand implements "git reset" style behavior
type ResetHardCommand struct{}

// Canonical name: matches Git
func (c *ResetHardCommand) Name() string { return "reset" }

// Usage string
func (c *ResetHardCommand) Usage() string { return "reset [--hard]" }

// Short description
func (c *ResetHardCommand) Description() string {
	return "Reset current branch state"
}

// Detailed description
func (c *ResetHardCommand) DetailedDescription() string {
	return `Reset the current branch.
Use --hard to discard all pending changes in the working directory (equivalent to 'git reset --hard').`
}

// Optional aliases
func (c *ResetHardCommand) Aliases() []string { return []string{"drop"} }

// One-letter shortcut
func (c *ResetHardCommand) Short() string { return "R" }

// Run executes the command
func (c *ResetHardCommand) Run(ctx *cli.Context) error {
	// parse flags
	hard := false
	for _, arg := range ctx.Args {
		switch arg {
		case "--hard":
			hard = true
		default:
			return fmt.Errorf("unknown option: %s", arg)
		}
	}

	if hard {
		return resetHard()
	}

	return fmt.Errorf("no reset mode specified (e.g., --hard)")
}

// resetHard discards all pending changes like 'git reset --hard'
func resetHard() error {
	branch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	lastCommitID, err := core.LastCommitID(branch.Name)
	if err != nil || lastCommitID == "" {
		return fmt.Errorf("no commit to reset to")
	}

	fmt.Printf("Resetting branch '%s' to last commit %s (--hard)...\n", branch.Name, lastCommitID)

	// Load last commit
	commitPath := filepath.Join(config.CommitsDir, lastCommitID+".json")
	var lastCommit core.Commit
	if err := util.ReadJSON(commitPath, &lastCommit); err != nil {
		return err
	}

	// Load associated fileset
	filesetPath := filepath.Join(config.FilesetsDir, lastCommit.FilesetID+".json")
	var fs snapshot.Fileset
	if err := util.ReadJSON(filesetPath, &fs); err != nil {
		return err
	}

	// Restore files
	if err := file.RestoreAll(fs.Files, fmt.Sprintf("reset --hard to %s", lastCommitID)); err != nil {
		return err
	}

	fmt.Println("Working directory successfully reset (--hard).")
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&ResetHardCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
