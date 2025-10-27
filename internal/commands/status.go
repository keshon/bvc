package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// StatusCommand shows uncommitted changes (added, modified, deleted)
type StatusCommand struct{}

// Canonical name
func (c *StatusCommand) Name() string { return "status" }

// Usage string
func (c *StatusCommand) Usage() string { return "status" }

// Short description
func (c *StatusCommand) Description() string {
	return "Show uncommitted changes"
}

// Detailed description
func (c *StatusCommand) DetailedDescription() string {
	return `List uncommitted changes in the current branch.
WARNING: Switching branches with pending changes may cause data loss.`
}

// Optional aliases
func (c *StatusCommand) Aliases() []string { return []string{"st"} }

// One-letter shortcut
func (c *StatusCommand) Short() string { return "S" }

// Run executes the command
func (c *StatusCommand) Run(ctx *cli.Context) error {
	return listPending()
}

// listPending calculates added, modified, and deleted files
func listPending() error {
	// Get current branch
	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

	// Load last commit's fileset
	commitID, _ := core.LastCommitID(currentBranch.Name)
	var lastFileset snapshot.Fileset
	if commitID != "" {
		var commit core.Commit
		commitPath := filepath.Join(config.CommitsDir, commitID+".json")
		if err := util.ReadJSON(commitPath, &commit); err == nil {
			fsPath := filepath.Join(config.FilesetsDir, commit.FilesetID+".json")
			_ = util.ReadJSON(fsPath, &lastFileset)
		}
	}

	lastFiles := make(map[string]file.Entry)
	for _, f := range lastFileset.Files {
		lastFiles[filepath.Clean(f.Path)] = f
	}

	// Build current workspace snapshot
	currFS, err := snapshot.Build()
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

// Register the command
func init() {
	cli.RegisterCommand(&StatusCommand{})
}
