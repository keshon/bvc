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

func (c *CommitCommand) Name() string        { return "commit" }
func (c *CommitCommand) Usage() string       { return "commit \"<message>\"" }
func (c *CommitCommand) Description() string { return "Commit current changes" }
func (c *CommitCommand) DetailedDescription() string {
	return "Commit changes to the current branch.\nMessage is mandatory."
}
func (c *CommitCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit message required")
	}
	message := ctx.Args[0]
	return commit(message)
}

func commit(message string) error {
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
	if bc, err := core.LastCommitID(branch.Name); err == nil {
		parent = bc
	}

	commitID := fmt.Sprintf("%x", time.Now().UnixNano())
	c := core.Commit{
		ID:        commitID,
		Parents:   []string{},
		Branch:    branch.Name,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: fileset.ID,
	}
	if parent != "" {
		c.Parents = append(c.Parents, parent)
	}

	if err := util.WriteJSON(filepath.Join(config.CommitsDir, commitID+".json"), c); err != nil {
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
