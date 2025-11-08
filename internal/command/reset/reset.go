package reset

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
)

type Command struct{}

func (c *Command) Name() string      { return "reset" }
func (c *Command) Short() string     { return "R" }
func (c *Command) Aliases() []string { return []string{"drop"} }
func (c *Command) Usage() string     { return "reset [<commit-id>] [--soft|--mixed|--hard]" }
func (c *Command) Brief() string     { return "Reset current branch to a commit or HEAD" }
func (c *Command) Help() string {
	return `Reset the current branch.
Modes:
  --soft  : move HEAD only
  --mixed : move HEAD and reset index (default)
  --hard  : move HEAD, reset index and working directory
If <commit-id> is omitted, the last commit is used (mixed).`
}

func (c *Command) Run(ctx *command.Context) error {
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

	// Open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	// If commit-id is not specified, use the last commit
	if targetID == "" {
		last, err := r.Meta.GetLastCommitID(branch.Name)
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

func reset(targetID, mode string) error {
	// Open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Load target commit
	target, err := r.Meta.GetCommit(targetID)
	if err != nil {
		return fmt.Errorf("unknown commit: %s", targetID)
	}

	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	fmt.Printf("Resetting branch '%s' to commit %s (%s)...\n", branch.Name, targetID, mode)

	switch mode {
	case "soft":
		// Move HEAD only
		if err := r.Meta.SetLastCommitID(branch.Name, targetID); err != nil {
			return err
		}

	case "mixed":
		// Move HEAD and reset index, keep working directory
		if err := r.Meta.SetLastCommitID(branch.Name, targetID); err != nil {
			return err
		}
		if err := resetIndex(target.FilesetID); err != nil {
			return err
		}

	case "hard":
		// Move HEAD, reset index and working directory
		if err := r.Meta.SetLastCommitID(branch.Name, targetID); err != nil {
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
	// Open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Load fileset
	fs, err := r.Store.Snapshots.Load(filesetID)
	if err != nil {
		return err
	}

	// Clear current staging and stage all files from the fileset
	if err := r.Store.Files.ClearIndex(); err != nil {
		return err
	}
	if err := r.Store.Files.StageFiles(fs.Files); err != nil {
		return err
	}

	fmt.Println("Index reset.")
	return nil
}

// resetWorkingDirectory restores files to the state of the commit
func resetWorkingDirectory(filesetID string) error {
	// Open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Load fileset
	fs, err := r.Store.Snapshots.Load(filesetID)
	if err != nil {
		return err
	}

	if err := r.Store.Files.Restore(fs.Files, fmt.Sprintf("reset --hard to fileset %s", filesetID)); err != nil {
		return err
	}

	fmt.Println("Working directory reset.")
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
