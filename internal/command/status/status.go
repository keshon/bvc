package status

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Command struct{}

func (c *Command) Name() string      { return "status" }
func (c *Command) Short() string     { return "S" }
func (c *Command) Aliases() []string { return []string{"st"} }
func (c *Command) Usage() string     { return "status" }
func (c *Command) Brief() string     { return "Show uncommitted changes" }
func (c *Command) Help() string {
	return `List uncommitted changes in the current branch.
WARNING: Switching branches with pending changes may cause data loss.`
}

func (c *Command) Run(ctx *command.Context) error {
	return status()
}

func status() error {
	// Open the repository context
	r, err := repo.OpenAt(config.DetectRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current branch
	currentBranch, err := r.GetCurrentBranch()
	if err != nil {
		return err
	}

	// Load last commit's fileset
	commitID, err := r.GetLastCommitID(currentBranch.Name)
	if err != nil {
		return err
	}

	var lastFileset snapshot.Fileset
	if commitID != "" {
		fs, err := r.GetCommitFileset(commitID)
		if err == nil {
			lastFileset = *fs
		}
	}

	lastFiles := make(map[string]file.Entry)
	for _, f := range lastFileset.Files {
		lastFiles[filepath.Clean(f.Path)] = f
	}

	// Create current snapshot
	currFS, err := r.Storage.Snapshots.CreateCurrent()
	if err != nil {
		return err
	}

	currFiles := make(map[string]file.Entry)
	for _, f := range currFS.Files {
		currFiles[filepath.Clean(f.Path)] = f
	}

	var added, modified, deleted []string

	// Detect added and modified
	for path, currFile := range currFiles {
		if lastFile, exists := lastFiles[path]; !exists {
			added = append(added, path)
		} else if !lastFile.Equal(&currFile) {
			modified = append(modified, path)
		}
	}

	// Detect deleted
	for path := range lastFiles {
		if _, exists := currFiles[path]; !exists {
			deleted = append(deleted, path)
		}
	}

	// Sort for consistent output
	sort.Strings(added)
	sort.Strings(modified)
	sort.Strings(deleted)

	// Display results
	fmt.Println("Pending changes:")
	if len(added) == 0 && len(modified) == 0 && len(deleted) == 0 {
		fmt.Println("  (no uncommitted changes)")
		return nil
	}

	printChanges("Added", "+", added)
	printChanges("Modified", "~", modified)
	printChanges("Deleted", "-", deleted)

	return nil
}

// printChanges prints a list of files with a prefix and gray-colored path
func printChanges(title, prefix string, files []string) {
	if len(files) == 0 {
		return
	}
	fmt.Println(title + ":")
	for _, f := range files {
		printPathWithGrayColor(prefix, f)
	}
	fmt.Print("\n")
}

// printPathWithGrayColor prints a single path with formatting
func printPathWithGrayColor(prefix, path string) {
	dir := filepath.Clean(filepath.Dir(path))
	if dir == "." || dir == string(os.PathSeparator) {
		dir = ""
	} else {
		dir += string(os.PathSeparator)
	}

	base := filepath.Base(path)
	full := dir + base

	const maxLen = 100
	if len(full) > maxLen {
		keep := maxLen - len(base) - 3
		if keep < 10 {
			keep = 10
		}
		if len(dir) > keep {
			start := dir[:keep/2]
			end := dir[len(dir)-(keep/2):]
			dir = start + "..." + end
		}
	}

	fmt.Printf("%s \033[90m%s\033[0m%s\n", prefix, dir, base)
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
