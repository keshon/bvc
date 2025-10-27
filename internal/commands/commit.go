package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/snapshot"

	"app/internal/util"
)

type CommitCommand struct{}

// Canonical name
func (c *CommitCommand) Name() string { return "commit" }

// Git-style usage
func (c *CommitCommand) Usage() string {
	return `commit -m "<message>" | --message="<message>"`
}

func (c *CommitCommand) Description() string {
	return "Commit current changes to the branch"
}

func (c *CommitCommand) DetailedDescription() string {
	return "Create a new commit with the staged changes.\nMessage is mandatory. Supports -m / --message flags or first positional argument."
}

// Short alias and one-letter
func (c *CommitCommand) Aliases() []string { return []string{"ci"} }
func (c *CommitCommand) Short() string     { return "c" }

// Run executes the commit
func (c *CommitCommand) Run(ctx *cli.Context) error {
	message := ""

	// Check flags first
	if val, ok := ctx.Flags["m"]; ok {
		message = val
	} else if val, ok := ctx.Flags["message"]; ok {
		message = val
	} else if len(ctx.Args) > 0 {
		// fallback to positional argument
		message = ctx.Args[0]
	}

	if message == "" {
		return fmt.Errorf("commit message required")
	}

	return c.commit(message)
}

// commit actualizes the commit
func (c *CommitCommand) commit(message string) error {
	fileset, err := snapshot.Build()
	if err != nil {
		return err
	}

	if err := fileset.Store(); err != nil {
		return err
	}

	filesetPath := filepath.Join(config.FilesetsDir, fileset.ID+".json")
	if err := util.WriteJSON(filesetPath, fileset); err != nil {
		return err
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

	fmt.Println("Committed:", commitID)
	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&CommitCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
