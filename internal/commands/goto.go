package commands

import (
	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/middleware"

	"app/internal/storage/file"
	"app/internal/storage/snapshot"

	"app/internal/util"
	"fmt"
	"os"
	"path/filepath"
)

type GotoCommand struct{}

func (c *GotoCommand) Name() string        { return "goto" }
func (c *GotoCommand) Usage() string       { return "goto <branch-name>" }
func (c *GotoCommand) Description() string { return "Switch to another branch" }
func (c *GotoCommand) DetailedDescription() string {
	return "Switch to another branch."
}
func (c *GotoCommand) Run(ctx *cli.Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("branch name required")
	}
	name := ctx.Args[0]
	return checkoutBranch(name)
}

func checkoutBranch(branch string) error {
	branchPath := filepath.Join(config.BranchesDir, branch)
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		return fmt.Errorf("branch does not exist")
	}

	commitIDBytes, err := os.ReadFile(branchPath)
	if err != nil {
		return err
	}
	commitID := string(commitIDBytes)

	if commitID == "" {
		emptyFS := snapshot.Fileset{ID: "empty", Files: nil}
		if err := file.RestoreAll(emptyFS.Files, fmt.Sprintf("empty branch '%s'", branch)); err != nil {
			return err
		}

		_, err = core.SetHeadRef(filepath.Join("branches", branch))
		if err != nil {
			return err
		}

		fmt.Println("Branch is empty, switched to", branch)
		return nil
	}

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

	if err := file.RestoreAll(fileset.Files, fmt.Sprintf("for branch '%s'", branch)); err != nil {
		return err
	}

	_, err = core.SetHeadRef(filepath.Join("branches", branch))
	if err != nil {
		return err
	}
	_ = core.SetLastCommit(branch, commitID)

	fmt.Println("Switched to branch", branch)
	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&GotoCommand{}, middleware.WithBlockIntegrityCheck()),
	)
}
