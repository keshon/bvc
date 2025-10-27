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

type PendingCommand struct{}

func (c *PendingCommand) Name() string        { return "pending" }
func (c *PendingCommand) Usage() string       { return "pending" }
func (c *PendingCommand) Description() string { return "Show uncommitted changes" }
func (c *PendingCommand) DetailedDescription() string {
	return "List uncommitted changes\nATTENTION: Dont switch branches having uncommitted changes, or (pending) files will be lost."
}
func (c *PendingCommand) Run(ctx *cli.Context) error {
	return listPending()
}

func listPending() error {
	currentBranch, err := core.CurrentBranch()
	if err != nil {
		return err
	}

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

	currFS, err := snapshot.Build()
	if err != nil {
		return err
	}

	currFiles := make(map[string]file.Entry)
	for _, f := range currFS.Files {
		currFiles[filepath.Clean(f.Path)] = f
	}

	var added, modified, deleted []string

	for path, currFile := range currFiles {
		if lastFile, exists := lastFiles[path]; !exists {
			added = append(added, path)
		} else if !lastFile.Equal(&currFile) {
			modified = append(modified, path)
		}
	}

	for path := range lastFiles {
		if _, exists := currFiles[path]; !exists {
			deleted = append(deleted, path)
		}
	}

	sort.Strings(added)
	sort.Strings(modified)
	sort.Strings(deleted)

	fmt.Println("Pending changes")
	if len(added) == 0 && len(modified) == 0 && len(deleted) == 0 {
		fmt.Println("  (no uncommitted changes)")
		return nil
	}

	if len(added) > 0 {
		fmt.Println("Added:")
		for _, f := range added {
			printPathWithGrayColor("+", f, 0)
		}
		fmt.Print("\n")
	}

	if len(modified) > 0 {
		fmt.Println("Modified:")
		for _, f := range modified {
			printPathWithGrayColor("~", f, 0)
		}
		fmt.Print("\n")
	}

	if len(deleted) > 0 {
		fmt.Println("Deleted:")
		for _, f := range deleted {
			printPathWithGrayColor("-", f, 0)
		}
	}

	return nil
}

func printPathWithGrayColor(prefix, path string, _ int) {
	dir := filepath.Clean(filepath.Dir(path))
	if dir == "." || dir == string(os.PathSeparator) {
		dir = ""
	} else {
		dir = dir + string(os.PathSeparator)
	}

	base := filepath.Base(path)
	full := dir + base

	const maxLen = 100
	if len(full) > maxLen {

		keep := maxLen - len(base) - 3
		if keep < 10 {
			keep = 10 // safety
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
	cli.RegisterCommand(&PendingCommand{})
}
