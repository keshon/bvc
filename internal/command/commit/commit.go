package commit

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repo/meta"
	"flag"
	"fmt"
	"strings"
	"time"
)

type Command struct{}

func (c *Command) Name() string  { return "commit" }
func (c *Command) Brief() string { return "Commit staged changes to the current branch" }
func (c *Command) Usage() string { return `commit -m "<message>" [--allow-empty]` }
func (c *Command) Help() string {
	return `Create a new commit with the staged changes.

Usage:
  commit -m "<message>"               - commit with a given message
  commit -m "<message>" --allow-empty - empty commit with a given message (no staged files exist)`
}
func (c *Command) Aliases() []string              { return []string{"ci"} }
func (c *Command) Subcommands() []command.Command { return nil }
func (c *Command) Flags(fs *flag.FlagSet)         {}

func (c *Command) Run(ctx *command.Context) error {
	var messages []string // commit messages
	var allowEmpty bool   // --allow-empty

	for i := 0; i < len(ctx.Args); i++ {
		arg := ctx.Args[i]

		switch {
		case arg == "-m" && i+1 < len(ctx.Args):
			messages = append(messages, ctx.Args[i+1])
			i++
		case strings.HasPrefix(arg, "-m="):
			messages = append(messages, strings.TrimPrefix(arg, "-m="))
		case arg == "--message" && i+1 < len(ctx.Args):
			messages = append(messages, ctx.Args[i+1])
			i++
		case strings.HasPrefix(arg, "--message="):
			messages = append(messages, strings.TrimPrefix(arg, "--message="))
		case arg == "--allow-empty":
			allowEmpty = true
		default:
			if len(messages) == 0 {
				messages = append(messages, arg)
			}
		}
	}

	if len(messages) == 0 {
		return fmt.Errorf("commit message required (use -m or pass message directly)")
	}

	message := strings.Join(messages, "\n\n")

	// open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// get staged files
	stagedFileentries, err := r.Store.Files.GetIndexFiles()
	if err != nil {
		return err
	}

	if len(stagedFileentries) == 0 && !allowEmpty {
		return fmt.Errorf("no staged changes to commit")
	}

	// create fileset from staged files (or empty fileset if --allow-empty)
	fileset, err := r.Store.Snapshots.Create(stagedFileentries)
	if err != nil {
		return err
	}

	if len(fileset.Files) > 0 {
		if err := r.Store.Snapshots.WriteAndSave(&fileset); err != nil {
			return err
		}
	}

	// create commit
	currentBranch, _ := r.Meta.GetCurrentBranch()
	parent := ""
	if last, err := r.Meta.GetLastCommitID(currentBranch.Name); err == nil {
		parent = last
	}

	newCommitID := fmt.Sprintf("%x", time.Now().UnixNano())
	newCommit := meta.Commit{
		ID:        newCommitID,
		Parents:   []string{},
		Branch:    currentBranch.Name,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: fileset.ID,
	}
	if parent != "" {
		newCommit.Parents = append(newCommit.Parents, parent)
	}

	_, err = r.Meta.CreateCommit(&newCommit)
	if err != nil {
		return err
	}
	if err := r.Meta.SetLastCommitID(currentBranch.Name, newCommitID); err != nil {
		return err
	}

	// clear staged changes after commit
	if len(stagedFileentries) > 0 {
		if err := r.Store.Files.ClearIndex(); err != nil {
			return err
		}
	}

	fmt.Println("Committed:", newCommitID)
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
