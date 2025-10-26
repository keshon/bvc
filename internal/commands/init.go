package commands

import (
	"app/internal/cli"
	"app/internal/core"
	"fmt"
	"os"
	"path/filepath"
)

type InitCommand struct{}

func (c *InitCommand) Name() string        { return "init" }
func (c *InitCommand) Usage() string       { return "init" }
func (c *InitCommand) Description() string { return "Initialize a new repository" }
func (c *InitCommand) DetailedDescription() string {
	return "Initialize a new repository in the current directory.\nIf the current directory is not empty, the content will be marked as pending."
}

func (c *InitCommand) Run(ctx *cli.Context) error {
	err := core.InitRepo()
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	parentDir := filepath.Base(wd)
	fmt.Printf("The repository \033[90m%s\033[0m has been initialized\n", parentDir)
	return nil
}

func init() {
	cli.RegisterCommand(&InitCommand{})
}
