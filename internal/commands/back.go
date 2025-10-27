package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/storage"
	"app/internal/util"
)

type BackCommand struct{}

func (c *BackCommand) Name() string        { return "back" }
func (c *BackCommand) Usage() string       { return "back <commit-id>" }
func (c *BackCommand) Description() string { return "Revert to a specific commit" }
func (c *BackCommand) DetailedDescription() string {
	return "Revert changes to a specific commit. Use log command to find the commit ID you want to revert to.\n"
}
func (c *BackCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit id required")
	}
	targetID := ctx.Args[0]
	return revertToCommit(targetID)
}

func revertToCommit(targetID string) error {
	commitPath := filepath.Join(config.CommitsDir, targetID+".json")
	var target core.Commit
	if err := util.ReadJSON(commitPath, &target); err != nil {
		return fmt.Errorf("unknown commit: %s", targetID)
	}

	fsPath := filepath.Join(config.FilesetsDir, target.FilesetID+".json")
	var fs storage.Fileset
	if err := util.ReadJSON(fsPath, &fs); err != nil {
		return err
	}

	fmt.Printf("Reverting to commit %s...\n", targetID)

	if err := storage.RestoreFileset(fs, fmt.Sprintf("for commit %s", targetID)); err != nil {
		return err
	}

	branch, _ := core.CurrentBranch()
	parent, _ := core.LastCommitID(branch.Name)
	newCommit := core.Commit{
		ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
		Parents:   []string{parent},
		Branch:    branch.Name,
		Message:   fmt.Sprintf("Revert to %s", targetID),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: target.FilesetID,
	}

	if err := util.WriteJSON(filepath.Join(config.CommitsDir, newCommit.ID+".json"), newCommit); err != nil {
		return err
	}
	if err := core.SetLastCommit(branch.Name, newCommit.ID); err != nil {
		return err
	}

	fmt.Println("Created revert commit:", newCommit.ID)
	fmt.Println("Workspace reverted to commit", targetID)
	return nil
}

func init() {
	cli.RegisterCommand(&BackCommand{})
}
