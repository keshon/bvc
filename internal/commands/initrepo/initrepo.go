package initrepo

import (
	"flag"
	"fmt"

	"bvc/internal/repo"
	"bvc/pkg/command"
)

type InitCmd struct {
	Path string
}

func (c *InitCmd) Name() string  { return "init" }
func (c *InitCmd) Brief() string { return "Initialize repository" }
func (c *InitCmd) Help() string  { return "Initialize repository" }
func (c *InitCmd) Flags(fs *flag.FlagSet) {
	fs.StringVar(&c.Path, "path", ".", "Path to initialize repository")
}
func (c *InitCmd) SubCommands() []command.Command { return nil }

func (c *InitCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(c.Path)
	if err != nil {
		return err
	}

	var mode string
	fmt.Println("Select repository mode (snapshot-first / stream-first):")
	fmt.Scanln(&mode)
	if mode != "snapshot-first" && mode != "stream-first" {
		mode = "snapshot-first"
	}

	r.Mode = mode

	if mode == "snapshot-first" {
		snap, err := r.CreateSnapshot("base", "initial")
		if err != nil {
			return err
		}
		r.HeadSetSnapshot(snap.ID)
	} else {
		if err := r.StreamCreate("main"); err != nil {
			return err
		}
		r.HeadSetStream("main")
	}

	fmt.Printf("Initialized repository at %s with mode %s\n", c.Path, mode)
	return nil
}
