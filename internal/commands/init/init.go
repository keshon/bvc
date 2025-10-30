package init

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
)

// Command initializes a new repository
type Command struct{}

// Canonical name
func (c *Command) Name() string { return "init" }

// Usage string
func (c *Command) Usage() string { return "init" }

// Short description
func (c *Command) Brief() string { return "Initialize a new repository" }

// Detailed description
func (c *Command) Help() string {
	return "Initialize a new repository in the current directory.\n" +
		"If the directory is not empty, existing content will be marked as pending changes."
}

// Aliases returns alternative names for the command
func (c *Command) Aliases() []string { return []string{"initialize"} }

// Short returns a single-letter shortcut
func (c *Command) Short() string { return "i" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
	repoName, err := core.InitRepo()
	if err != nil {
		return err
	}

	fmt.Printf("Repository \033[90m%s\033[0m has been initialized\n", repoName)
	return nil
}

// Register the command
func init() {
	cli.RegisterCommand(&Command{})
}
