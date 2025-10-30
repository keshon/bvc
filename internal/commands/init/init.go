package init

import (
	"fmt"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
)

type Command struct{}

func (c *Command) Name() string      { return "init" }
func (c *Command) Short() string     { return "i" }
func (c *Command) Aliases() []string { return []string{"initialize"} }
func (c *Command) Usage() string     { return "init" }
func (c *Command) Brief() string     { return "Initialize a new repository" }
func (c *Command) Help() string {
	return "Initialize a new repository in the current directory.\n" +
		"If the directory is not empty, existing content will be marked as pending changes."
}

func (c *Command) Run(ctx *cli.Context) error {
	exists, err := core.RepoExists()
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("repository already exists")
	}

	name, err := core.InitRepo()
	if err != nil {
		return err
	}

	fmt.Printf("Repository \033[90m%s\033[0m has been initialized\n", name)
	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
