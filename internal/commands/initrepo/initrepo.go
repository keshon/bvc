package initrepo

import (
	"bvc/command"
	"bvc/internal/util"
	"flag"
	"fmt"
)

type InitCmd struct {
	Path string
}

func (c *InitCmd) Name() string { return "init" }
func (c *InitCmd) Help() string { return "Initialize repository" }
func (c *InitCmd) Flags(fs *flag.FlagSet) {
	fs.StringVar(&c.Path, "path", ".", "Path to initialize repository")
}
func (c *InitCmd) SubCommands() []command.Command { return nil }

func (c *InitCmd) Run(ctx command.Context) error {
	repo, err := util.OpenRepo(c.Path)
	if err != nil {
		return err
	}

	var mode string
	fmt.Println("Select repository mode (snapshot-first / stream-first):")
	fmt.Scanln(&mode)
	if mode != "snapshot-first" && mode != "stream-first" {
		mode = "snapshot-first"
	}

	repo.Mode = mode

	if mode == "snapshot-first" {
		snap, err := repo.CreateSnapshot("base", "initial")
		if err != nil {
			return err
		}
		repo.HeadSetSnapshot(snap.ID)
	} else {
		if err := repo.StreamCreate("main"); err != nil {
			return err
		}
		repo.HeadSetStream("main")
	}

	fmt.Printf("Initialized repository at %s with mode %s\n", c.Path, mode)
	return nil
}
