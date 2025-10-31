package commit

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
	"fmt"
	"strings"
	"time"
)

type Command struct{}

func (c *Command) Name() string      { return "commit" }
func (c *Command) Short() string     { return "c" }
func (c *Command) Aliases() []string { return []string{"ci"} }
func (c *Command) Usage() string     { return `commit -m "<message>" [--allow-empty]` }
func (c *Command) Brief() string     { return "Commit staged changes to the current branch" }
func (c *Command) Help() string {
	return `Create a new commit with the staged changes.
Supports -m / --message for commit message.
Supports --allow-empty to commit even if no staged changes exist.`
}

func (c *Command) Run(ctx *command.Context) error {
	var messages []string
	var allowEmpty bool

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
			// fallback: if no -m given, treat arg as commit message
			if len(messages) == 0 {
				messages = append(messages, arg)
			}
		}
	}

	if len(messages) == 0 {
		return fmt.Errorf("commit message required (use -m or pass message directly)")
	}

	message := strings.Join(messages, "\n\n")
	return commitChanges(message, allowEmpty)
}

func commitChanges(message string, allowEmpty bool) error {
	// Open the repository context
	r, err := core.OpenAt(config.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get staged files
	stagedFileentries, err := file.GetIndexFiles()
	if err != nil {
		return err
	}

	if len(stagedFileentries) == 0 && !allowEmpty {
		return fmt.Errorf("no staged changes to commit")
	}

	// Create fileset from staged files (empty fileset allowed with --allow-empty)
	fileset, err := snapshot.CreateFileset(stagedFileentries)
	if err != nil {
		return err
	}

	if len(fileset.Files) > 0 {
		if err := fileset.WriteAndSaveFileset(); err != nil {
			return err
		}
	}

	branch, _ := r.GetCurrentBranch()
	parent := ""
	if last, err := r.GetLastCommitID(branch.Name); err == nil {
		parent = last
	}

	commitID := fmt.Sprintf("%x", time.Now().UnixNano())
	cmt := core.Commit{
		ID:        commitID,
		Parents:   []string{},
		Branch:    branch.Name,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: fileset.ID,
	}
	if parent != "" {
		cmt.Parents = append(cmt.Parents, parent)
	}

	_, err = r.CreateCommit(&cmt)
	if err != nil {
		return err
	}
	if err := r.SetLastCommitID(branch.Name, commitID); err != nil {
		return err
	}

	// Clear staged changes after commit
	if len(stagedFileentries) > 0 {
		if err := file.ClearIndex(); err != nil {
			return err
		}
	}

	fmt.Println("Committed:", commitID)
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
