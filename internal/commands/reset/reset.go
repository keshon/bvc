package reset

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
)

// Command implements Git-like reset
type Command struct{}

// Name
func (c *Command) Name() string { return "reset" }

// Usage
func (c *Command) Usage() string { return "reset [<commit-id>] [--soft|--mixed|--hard]" }

// Description
func (c *Command) Brief() string { return "Reset current branch to a commit or HEAD" }

// Detailed description
func (c *Command) Help() string {
	return `Reset the current branch.
Modes:
  --soft  : move HEAD only
  --mixed : move HEAD and reset index (default)
  --hard  : move HEAD, reset index and working directory
If <commit-id> is omitted, the last commit is used (mixed).`
}

// Aliases
func (c *Command) Aliases() []string { return []string{"drop"} }

// Shortcut
func (c *Command) Short() string { return "R" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	var targetID string
	mode := "mixed"
	var modeSet bool

	for _, arg := range ctx.Args {
		switch arg {
		case "--soft":
			if modeSet && mode != "soft" {
				return fmt.Errorf("conflicting reset modes: %s and soft", mode)
			}
			mode = "soft"
			modeSet = true
		case "--mixed":
			if modeSet && mode != "mixed" {
				return fmt.Errorf("conflicting reset modes: %s and mixed", mode)
			}
			mode = "mixed"
			modeSet = true
		case "--hard":
			if modeSet && mode != "hard" {
				return fmt.Errorf("conflicting reset modes: %s and hard", mode)
			}
			mode = "hard"
			modeSet = true
		default:
			if targetID == "" {
				targetID = arg
			} else {
				return fmt.Errorf("unknown option or duplicate commit-id: %s", arg)
			}
		}
	}

	branch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	// If commit-id is not specified, use the last commit
	if targetID == "" {
		last, err := core.LastCommitID(branch.Name)
		if err != nil {
			return fmt.Errorf("cannot determine last commit: %v", err)
		}
		if last == "" {
			return fmt.Errorf("no commits to reset to")
		}
		targetID = last
	}

	return reset(targetID, mode)
}

// reset performs the actual reset based on mode
func reset(targetID, mode string) error {
	// Load target commit
	target, err := core.GetCommit(targetID)
	if err != nil {
		return fmt.Errorf("unknown commit: %s", targetID)
	}

	branch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	fmt.Printf("Resetting branch '%s' to commit %s (%s)...\n", branch.Name, targetID, mode)

	switch mode {
	case "soft":
		// Move HEAD only
		if err := core.SetLastCommitID(branch.Name, targetID); err != nil {
			return err
		}

	case "mixed":
		// Move HEAD and reset index, keep working directory
		if err := core.SetLastCommitID(branch.Name, targetID); err != nil {
			return err
		}
		if err := resetIndex(target.FilesetID); err != nil {
			return err
		}

	case "hard":
		// Move HEAD, reset index and working directory
		if err := core.SetLastCommitID(branch.Name, targetID); err != nil {
			return err
		}
		if err := resetIndex(target.FilesetID); err != nil {
			return err
		}
		if err := resetWorkingDirectory(target.FilesetID); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported reset mode: %s", mode)
	}

	fmt.Println("Reset complete.")
	return nil
}

// resetIndex resets the staging area to the specified fileset
func resetIndex(filesetID string) error {
	// Load fileset
	fs, err := snapshot.LoadFileset(filesetID)
	if err != nil {
		return err
	}

	// Clear current staging and stage all files from the fileset
	if err := file.ClearIndex(); err != nil {
		return err
	}
	if err := file.StageFiles(fs.Files); err != nil {
		return err
	}

	fmt.Println("Index reset.")
	return nil
}

// resetWorkingDirectory restores files to the state of the commit
func resetWorkingDirectory(filesetID string) error {
	// Load fileset
	fs, err := snapshot.LoadFileset(filesetID)
	if err != nil {
		return err
	}

	if err := file.RestoreFiles(fs.Files, fmt.Sprintf("reset --hard to fileset %s", filesetID)); err != nil {
		return err
	}

	fmt.Println("Working directory reset.")
	return nil
}

// Register command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&Command{}, middleware.WithBlockIntegrityCheck()),
	)
}
