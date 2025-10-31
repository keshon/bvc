package init

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
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

func (c *Command) Run(ctx *command.Context) error {
	repo, created, err := repo.InitAt(config.RepoDir)
	if err != nil {
		return err
	}
	rootDir := repo.Root()

	if created {
		fmt.Printf("Repository %q has been initialized\n", rootDir)
	} else {
		fmt.Printf("Repository %q already initialized\n", rootDir)
	}
	return nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
