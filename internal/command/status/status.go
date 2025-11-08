package status

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repo/store/file"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Command struct{}

func (c *Command) Name() string      { return "status" }
func (c *Command) Short() string     { return "S" }
func (c *Command) Aliases() []string { return []string{"st"} }
func (c *Command) Usage() string     { return "status [options]" }
func (c *Command) Brief() string     { return "Show uncommitted changes" }
func (c *Command) Help() string {
	return `Show the working tree status.

Options:
  -s, --short                    Show short summary (one line per file)
  -b, --branch                   Show branch info
  -u, --untracked-files=<mode>   Show untracked files: no, normal, all (default: normal)
      --ignored                  Show ignored files
  -q, --quiet                    Suppress normal output

Examples:
  bvc status
  bvc status -s
  bvc status --branch
  bvc status -u all
  bvc status --ignored
`
}

func (c *Command) Run(ctx *command.Context) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)

	short := fs.Bool("short", false, "")
	fs.BoolVar(short, "s", false, "alias for --short")

	branch := fs.Bool("branch", false, "")
	fs.BoolVar(branch, "b", false, "alias for --branch")

	untracked := fs.String("untracked-files", "normal", "")
	fs.StringVar(untracked, "u", "normal", "alias for --untracked-files")

	ignored := fs.Bool("ignored", false, "")
	quiet := fs.Bool("quiet", false, "")
	fs.BoolVar(quiet, "q", false, "alias for --quiet")

	if err := fs.Parse(ctx.Args); err != nil {
		return err
	}

	return status(*short, *branch, *untracked, *ignored, *quiet)
}

// status performs the main status logic
func status(short, showBranch bool, untrackedMode string, showIgnored, quiet bool) error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Branch info
	currentBranch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		return err
	}
	if showBranch && !quiet {
		fmt.Printf("On branch %s\n", currentBranch.Name)
	}

	// Load last commit snapshot
	lastFiles := map[string]file.Entry{}
	if commitID, err := r.Meta.GetLastCommitID(currentBranch.Name); err == nil && commitID != "" {
		if fs, err := r.GetCommitFileset(commitID); err == nil {
			for _, f := range fs.Files {
				lastFiles[filepath.Clean(f.Path)] = f
			}
		}
	}

	// Load current snapshot
	currFS, err := r.Store.Snapshots.CreateCurrent()
	if err != nil {
		return err
	}

	currFiles := map[string]file.Entry{}
	for _, f := range currFS.Files {
		currFiles[filepath.Clean(f.Path)] = f
	}

	// Detect changes
	added, modified, deleted := detectChanges(lastFiles, currFiles)

	if quiet {
		return nil
	}

	if short {
		printShort(added, modified, deleted, untrackedMode, showIgnored)
	} else {
		printFull(added, modified, deleted, untrackedMode, showIgnored)
	}

	return nil
}

// detectChanges compares last and current filesets
func detectChanges(lastFiles, currFiles map[string]file.Entry) (added, modified, deleted []string) {
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

	// deterministic output
	sort.Strings(added)
	sort.Strings(modified)
	sort.Strings(deleted)

	return
}

// printFull prints the normal status output
func printFull(added, modified, deleted []string, untrackedMode string, showIgnored bool) {
	fmt.Println("Pending changes:")
	printChanges("Added", "+", added)
	printChanges("Modified", "~", modified)
	printChanges("Deleted", "-", deleted)

	if untrackedMode != "no" {
		fmt.Println("Untracked files: (not implemented yet)")
	}
	if showIgnored {
		fmt.Println("Ignored files: (not implemented yet)")
	}
}

// printShort prints a deterministic one-line-per-file summary
func printShort(added, modified, deleted []string, untrackedMode string, showIgnored bool) {
	type entry struct {
		path, status string
	}

	var all []entry
	for _, f := range added {
		all = append(all, entry{f, "A"})
	}
	for _, f := range modified {
		all = append(all, entry{f, "M"})
	}
	for _, f := range deleted {
		all = append(all, entry{f, "D"})
	}

	// deterministic sort by path
	sort.Slice(all, func(i, j int) bool { return all[i].path < all[j].path })

	for _, e := range all {
		printPathWithGrayColor(e.status+" ", e.path)
	}

	if untrackedMode != "no" {
		fmt.Println("(untracked files detection not implemented yet)")
	}
	if showIgnored {
		fmt.Println("(ignored files detection not implemented yet)")
	}
}

// printChanges prints a section of files with a prefix
func printChanges(title, prefix string, files []string) {
	if len(files) == 0 {
		return
	}
	fmt.Println(title + ":")
	for _, f := range files {
		printPathWithGrayColor(prefix, f)
	}
	fmt.Println()
}

// printPathWithGrayColor prints a file path with gray-colored directory
func printPathWithGrayColor(prefix, path string) {
	dir := filepath.Clean(filepath.Dir(path))
	if dir == "." || dir == string(os.PathSeparator) {
		dir = ""
	} else {
		dir += string(os.PathSeparator)
	}

	base := filepath.Base(path)
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
