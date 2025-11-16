package cherry_pick

import (
	"flag"
	"fmt"
	"time"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/internal/repo/meta"
)

type Command struct{}

func (c *Command) Name() string  { return "cherry-pick" }
func (c *Command) Brief() string { return "Apply selected commit to the current branch" }
func (c *Command) Usage() string { return "cherry-pick <commit-id>" }
func (c *Command) Help() string {
	return `Apply a specific commit to the current branch.

Usage:
  cherry-pick <commit-id>`
}
func (c *Command) Aliases() []string              { return []string{"cp"} }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet)         {}

func (c *Command) Run(ctx *command.Context) error {
	// require commit ID
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit ID required")
	}
	commitID := ctx.Args[0]

	// open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// get commit and fileset
	targetCommit, err := r.Meta.GetCommit(commitID)
	if err != nil {
		return err
	}

	targetFileset, err := r.Store.Snapshots.Load(targetCommit.FilesetID)
	if err != nil {
		return err
	}

	// get current branch
	targetBranch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}

	// get parent commit
	parent, err := r.Meta.GetLastCommitID(targetBranch.Name)
	if err != nil {
		return err
	}

	// create new commit on current branch referencing the picked commit
	newCommit := meta.Commit{
		ID:        fmt.Sprintf("%x", time.Now().UnixNano()),
		Parents:   []string{parent},
		Branch:    targetBranch.Name,
		Message:   fmt.Sprintf("Pick commit %s", commitID),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: targetCommit.FilesetID,
	}

	// create commit
	_, err = r.Meta.CreateCommit(&newCommit)
	if err != nil {
		return err
	}

	// update last commit for the branch
	if err := r.Meta.SetLastCommitID(targetBranch.Name, newCommit.ID); err != nil {
		return err
	}

	// restore files from picked commit
	if err := r.Store.Files.RestoreFilesToWorkingTree(targetFileset.Files, fmt.Sprintf("pick commit %s", commitID)); err != nil {
		return err
	}

	fmt.Printf("Picked commit %s into branch '%s' as %s\n", commitID, targetBranch.Name, newCommit.ID)
	return nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
			middleware.WithBlockIntegrityCheck(),
		),
	)
}
