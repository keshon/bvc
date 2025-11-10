package reset

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"flag"
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

func (c *Command) Subcommands() []command.Command {
	return nil
}
func (c *Command) Flags(fs *flag.FlagSet) {
	fs.Bool("soft", false, "move HEAD only")
	fs.Bool("mixed", false, "move HEAD and reset index (default)")
	fs.Bool("hard", false, "move HEAD, reset index and working directory")
}

// Run executes the reset command
func (c *Command) Run(ctx *command.Context) error {
	// extract flags
	soft := ctx.Flags.Lookup("soft").Value.(flag.Getter).Get().(bool)
	hard := ctx.Flags.Lookup("hard").Value.(flag.Getter).Get().(bool)

	mode := "mixed"
	if soft {
		mode = "soft"
	}
	if hard {
		if mode != "mixed" {
			return fmt.Errorf("conflicting reset modes: %s and hard", mode)
		}
		mode = "hard"
	}

	// extract optional commit-id (non-flag args)
	targetID := ""
	if len(ctx.Args) > 0 {
		targetID = ctx.Args[0]
	}

	// Open repository
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	// If no commit-id, use last commit
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
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

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
		if err := r.Meta.SetLastCommitID(branch.Name, targetID); err != nil {
			return err
		}
	case "mixed":
		if err := r.Meta.SetLastCommitID(branch.Name, targetID); err != nil {
			return err
		}
		if err := resetIndex(target.FilesetID); err != nil {
			return err
		}
	case "hard":
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

func resetIndex(filesetID string) error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	fs, err := r.Store.Snapshots.Load(filesetID)
	if err != nil {
		return err
	}

	if err := r.Store.Files.ClearIndex(); err != nil {
		return err
	}
	if err := r.Store.Files.SaveIndex(fs.Files); err != nil {
		return err
	}

	fmt.Println("Index reset.")
	return nil
}

func resetWorkingDirectory(filesetID string) error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	fs, err := r.Store.Snapshots.Load(filesetID)
	if err != nil {
		return err
	}

	if err := r.Store.Files.RestoreFilesToWorkingTree(fs.Files, fmt.Sprintf("reset --hard to fileset %s", filesetID)); err != nil {
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
