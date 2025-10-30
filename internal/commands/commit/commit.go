package commit

import (
	"fmt"
	"strings"
	"time"

	"app/internal/cli"
	"app/internal/core"
	"app/internal/middleware"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
)

// Command implements Git-like commit behavior
type Command struct{}

// Name
func (c *Command) Name() string { return "commit" }

// Usage
func (c *Command) Usage() string { return `commit -m "<message>" [--allow-empty]` }

// Short description
func (c *Command) Brief() string {
	return "Commit staged changes to the current branch"
}

// Detailed description
func (c *Command) Help() string {
	return `Create a new commit with the staged changes.
Supports -m / --message for commit message.
Supports --allow-empty to commit even if no staged changes exist.`
}

// Optional aliases
func (c *Command) Aliases() []string { return []string{"ci"} }

// One-letter shortcut
func (c *Command) Short() string { return "c" }

// Run executes the command
func (c *Command) Run(ctx *cli.Context) error {
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
	return c.commit(message, allowEmpty)
}

// commit actualizes a new commit
func (c *Command) commit(message string, allowEmpty bool) error {
	// Get staged files
	stagedFileentries, err := file.GetIndexFiles()
	if err != nil {
		return err
	}

	if len(stagedFileentries) == 0 && !allowEmpty {
		return fmt.Errorf("no staged changes to commit")
	}

	// Create fileset from staged files (empty fileset allowed with --allow-empty)
	fileset, err := snapshot.CreateFilesetFromEntries(stagedFileentries)
	if err != nil {
		return err
	}

	if len(fileset.Files) > 0 {
		if err := fileset.WriteAndSaveFileset(); err != nil {
			return err
		}
	}

	branch, _ := core.CurrentBranch()
	parent := ""
	if last, err := core.LastCommitID(branch.Name); err == nil {
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

	_, err = core.CreateCommit(&cmt)
	if err != nil {
		return err
	}
	if err := core.SetLastCommitID(branch.Name, commitID); err != nil {
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

// Register the command
func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(&Command{}, middleware.WithBlockIntegrityCheck()),
	)
}
