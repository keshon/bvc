package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"app/internal/cli"
	"app/internal/core"
)

// InitCommand initializes a new repository
type InitCommand struct{}

// Canonical name
func (c *InitCommand) Name() string { return "init" }

// Usage string
func (c *InitCommand) Usage() string { return "init" }

// Short description
func (c *InitCommand) Description() string { return "Initialize a new repository" }

// Detailed description
func (c *InitCommand) DetailedDescription() string {
	return "Initialize a new repository in the current directory.\n" +
		"If the directory is not empty, existing content will be marked as pending changes."
}

// Aliases returns alternative names for the command
func (c *InitCommand) Aliases() []string { return []string{"initialize"} }

// Short returns a single-letter shortcut
func (c *InitCommand) Short() string { return "i" }

// Run executes the command
func (c *InitCommand) Run(ctx *cli.Context) error {
	// Initialize repository structure
	if err := core.InitRepo(); err != nil {
		return err
	}

	// Get current directory name
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoName := filepath.Base(wd)

	fmt.Printf("Repository \033[90m%s\033[0m has been initialized\n", repoName)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(&InitCommand{})
}
