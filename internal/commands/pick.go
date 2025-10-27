package commands

import (
	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage"
	"app/internal/util"
	"fmt"
	"path/filepath"
	"time"
)

type PickCommand struct{}

func (c *PickCommand) Name() string        { return "pick" }
func (c *PickCommand) Usage() string       { return "pick <commit-id>" }
func (c *PickCommand) Description() string { return "Apply selected commit to current branch" }
func (c *PickCommand) DetailedDescription() string {
	return "Apply a specific commit to current branch\nUse 'bvc log all' command to find the commit ID you want to grab"
}

func (c *PickCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit ID required")
	}
	commitID := ctx.Args[0]
	return pickCommit(commitID)
}

func pickCommit(commitID string) error {
	targetPath := filepath.Join(config.CommitsDir, commitID+".json")
	var target core.Commit
	if err := util.ReadJSON(targetPath, &target); err != nil {
		return fmt.Errorf("unknown commit: %s", commitID)
	}

	fsPath := filepath.Join(config.FilesetsDir, target.FilesetID+".json")
	var fs storage.Fileset
	if err := util.ReadJSON(fsPath, &fs); err != nil {
		return err
	}

	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}
	parent, err := core.LastCommitID(currentBranch.Name)
	if err != nil {
		return err
	}

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

	if err := core.SetLastCommit(currentBranch.Name, newCommit.ID); err != nil {
		return err
	}

	if err := storage.RestoreFileset(fs, fmt.Sprintf("pick commit %s", commitID)); err != nil {
		return err
	}

	fmt.Printf("Picked commit %s into branch '%s' as %s\n", commitID, currentBranch.Name, newCommit.ID)
	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&PickCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
