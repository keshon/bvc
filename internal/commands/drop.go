package commands

import (
	"fmt"
	"path/filepath"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"

	"app/internal/storage/file"
	"app/internal/storage/snapshot"

	"app/internal/util"
)

type DropCommand struct{}

func (c *DropCommand) Name() string        { return "drop" }
func (c *DropCommand) Usage() string       { return "drop" }
func (c *DropCommand) Description() string { return "Discard pending changes" }
func (c *DropCommand) DetailedDescription() string {
	return "Discard pending changes to the last commit"
}
func (c *DropCommand) Run(ctx *cli.Context) error {
	return discardChanges()
}

func discardChanges() error {
	branch, err := core.CurrentBranch()
	if err != nil {
		return err
	}
	commitID, err := core.LastCommitID(branch.Name)
	if err != nil || commitID == "" {
		return fmt.Errorf("no commit to drop to")
	}

	fmt.Printf("Discarding changes and restoring branch '%s' to commit %s...\n", branch, commitID)

	commitPath := filepath.Join(config.CommitsDir, commitID+".json")
	var c core.Commit
	if err := util.ReadJSON(commitPath, &c); err != nil {
		return err
	}

	fsPath := filepath.Join(config.FilesetsDir, c.FilesetID+".json")
	var fileset snapshot.Fileset
	if err := util.ReadJSON(fsPath, &fileset); err != nil {
		return err
	}

	if err := file.RestoreAll(fileset.Files, fmt.Sprintf("discard to %s", commitID)); err != nil {
		return err
	}

	fmt.Println("Working directory restored to last commit.")
	return nil
}

func init() {
	cli.RegisterCommand(&DropCommand{})
}
