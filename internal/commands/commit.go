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

// CommitCommand implements Git-like commit behavior
type CommitCommand struct{}

func (c *CommitCommand) Name() string  { return "commit" }
func (c *CommitCommand) Usage() string { return `commit -m "<message>" [--allow-empty]` }
func (c *CommitCommand) Description() string {
	return "Commit staged changes to the current branch"
}
func (c *CommitCommand) DetailedDescription() string {
	return `Create a new commit with the staged changes.
Supports -m / --message for commit message.
Supports --allow-empty to commit even if no staged changes exist.`
}
func (c *CommitCommand) Aliases() []string { return []string{"ci"} }
func (c *CommitCommand) Short() string     { return "c" }

// Run executes the commit command
func (c *CommitCommand) Run(ctx *cli.Context) error {
	message := ""

	// Check flags first
	if val, ok := ctx.Flags["m"]; ok {
		message = val
	} else if val, ok := ctx.Flags["message"]; ok {
		message = val
	} else if len(ctx.Args) > 0 {
		// fallback to first positional argument
		message = ctx.Args[0]
	}

	if message == "" {
		return fmt.Errorf("commit message required")
	}

	allowEmpty := false
	if _, ok := ctx.Flags["allow-empty"]; ok {
		allowEmpty = true
	}

	return c.commit(message, allowEmpty)
}

// commit actualizes a new commit
func (c *CommitCommand) commit(message string, allowEmpty bool) error {
	// Get staged files
	stagedFiles, err := file.GetIndexFiles()
	if err != nil {
		return err
	}

	if len(stagedFiles) == 0 && !allowEmpty {
		return fmt.Errorf("no staged changes to commit")
	}

	// Build fileset from staged files (empty fileset allowed with --allow-empty)
	fileset, err := snapshot.BuildFromFiles(stagedFiles)
	if err != nil {
		return err
	}

	if len(fileset.Files) > 0 {
		if err := fileset.Store(); err != nil {
			return err
		}

		filesetPath := filepath.Join(config.FilesetsDir, fileset.ID+".json")
		if err := util.WriteJSON(filesetPath, fileset); err != nil {
			return err
		}
	}

	branch, _ := core.CurrentBranch()
	parent := ""
	if last, err := core.LastCommitID(branch.Name); err == nil {
		parent = last
	}

	commitID := fmt.Sprintf("%x", time.Now().UnixNano())
	cmt := core.Commit{
		ID:        commitID,
		Parents:   []string{},
		Branch:    branch.Name,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: fileset.ID,
	}
	if parent != "" {
		cmt.Parents = append(cmt.Parents, parent)
	}

	if err := util.WriteJSON(filepath.Join(config.CommitsDir, commitID+".json"), cmt); err != nil {
		return err
	}
	if err := core.SetLastCommit(branch.Name, commitID); err != nil {
		return err
	}

	// Clear staged changes after commit
	if len(stagedFiles) > 0 {
		if err := file.ClearIndex(); err != nil {
			return err
		}
	}

	fmt.Println("Committed:", commitID)
	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&CommitCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
