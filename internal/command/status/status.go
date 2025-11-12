package status

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repo/store/file"
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Command struct{}

func (c *Command) Name() string      { return "status" }
func (c *Command) Short() string     { return "S" }
func (c *Command) Aliases() []string { return []string{"st"} }
func (c *Command) Usage() string     { return "status [options]" }
func (c *Command) Brief() string     { return "Show working tree and index status" }

func (c *Command) Help() string {
	return `Show the working tree status.

Options:
  -s, --short                    Show short summary (XY path)
      --porcelain                Machine-readable short output
  -b, --branch                   Show branch info
  -u, --untracked-files=<mode>   Show untracked files: no, normal, all (default: normal)
      --ignored                  Show ignored files
  -q, --quiet                    Suppress normal output

Notes:
  "porcelain" mode is a stable, machine-readable short output format (like -s).
`
}

func (c *Command) Subcommands() []command.Command { return nil }

func (c *Command) Flags(fs *flag.FlagSet) {
	fs.Bool("short", false, "show short summary")
	fs.Bool("s", false, "alias for --short")
	fs.Bool("porcelain", false, "machine-readable short output")
	fs.Bool("branch", false, "")
	fs.Bool("b", false, "alias for --branch")
	fs.String("untracked-files", "normal", "")
	fs.String("u", "normal", "alias for --untracked-files")
	fs.Bool("ignored", false, "")
	fs.Bool("quiet", false, "")
	fs.Bool("q", false, "alias for --quiet")
}

func (c *Command) Run(ctx *command.Context) error {
	short := ctx.Flags.Lookup("short").Value.(flag.Getter).Get().(bool) ||
		ctx.Flags.Lookup("s").Value.(flag.Getter).Get().(bool)
	porcelain := ctx.Flags.Lookup("porcelain").Value.(flag.Getter).Get().(bool)
	showBranch := ctx.Flags.Lookup("branch").Value.(flag.Getter).Get().(bool) ||
		ctx.Flags.Lookup("b").Value.(flag.Getter).Get().(bool)
	untrackedMode := ctx.Flags.Lookup("untracked-files").Value.(flag.Getter).Get().(string)
	if u := ctx.Flags.Lookup("u"); u != nil {
		untrackedMode = u.Value.(flag.Getter).Get().(string)
	}
	showIgnored := ctx.Flags.Lookup("ignored").Value.(flag.Getter).Get().(bool)
	quiet := ctx.Flags.Lookup("quiet").Value.(flag.Getter).Get().(bool) ||
		ctx.Flags.Lookup("q").Value.(flag.Getter).Get().(bool)

	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		if strings.Contains(err.Error(), "HEAD") {
			if !quiet {
				fmt.Println("No commits yet on any branch")
			}
			return nil
		}
		return err
	}

	// HEAD files
	headFiles := map[string]file.Entry{}
	if commitID, err := r.Meta.GetLastCommitID(branch.Name); err == nil && commitID != "" {
		fs, err := r.GetCommitFileset(commitID)
		if err != nil {
			return err
		}
		for _, e := range fs.Files {
			headFiles[filepath.Clean(e.Path)] = e
		}
	}

	// index
	indexEntries, _ := r.Store.Files.LoadIndex()
	indexFiles := map[string]file.Entry{}
	for _, e := range indexEntries {
		indexFiles[filepath.Clean(e.Path)] = e
	}

	// working tree (tracked + ignored separated)
	workFS, ignoredFS, err := r.Store.Snapshots.BuildFilesetFromWorkingTree()
	if err != nil {
		return fmt.Errorf("scan working tree: %w", err)
	}

	workFiles := make(map[string]file.Entry, len(workFS.Files))
	for _, e := range workFS.Files {
		workFiles[filepath.Clean(e.Path)] = e
	}

	ignoredFiles := make(map[string]file.Entry, len(ignoredFS.Files))
	for _, e := range ignoredFS.Files {
		ignoredFiles[filepath.Clean(e.Path)] = e
	}

	// collect all paths from HEAD, index, and work (not ignored)
	// Only tracked files
	allPaths := map[string]struct{}{}
	for p := range headFiles {
		allPaths[p] = struct{}{}
	}
	for p := range indexFiles {
		allPaths[p] = struct{}{}
	}
	for p := range workFiles { // tracked work files only
		allPaths[p] = struct{}{}
	}
	paths := make([]string, 0, len(allPaths))
	for p := range allPaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var (
		stagedAdded, stagedModified, stagedDeleted []string
		unstagedModified, unstagedDeleted          []string
		untracked, ignored                         []string
	)

	// diff logic
	for _, p := range paths {
		h, inHead := headFiles[p]
		i, inIndex := indexFiles[p]
		w, inWork := workFiles[p]

		// staged
		if inIndex {
			if !inHead {
				stagedAdded = append(stagedAdded, p)
			} else if len(i.Blocks) == 0 {
				stagedDeleted = append(stagedDeleted, p)
			} else if !h.Equal(&i) {
				stagedModified = append(stagedModified, p)
			}
		}

		// unstaged
		if inWork {
			if inIndex {
				if !i.Equal(&w) {
					unstagedModified = append(unstagedModified, p)
				}
			} else if inHead {
				if !h.Equal(&w) {
					unstagedModified = append(unstagedModified, p)
				}
			} else if untrackedMode != "no" {
				untracked = append(untracked, p)
			}
		} else if inHead && !inIndex {
			unstagedDeleted = append(unstagedDeleted, p)
		}
	}

	// ignored list
	if showIgnored {
		for _, e := range ignoredFS.Files {
			ignored = append(ignored, e.Path)
		}
	}

	if quiet {
		return nil
	}

	// output
	if showBranch {
		fmt.Printf("On branch %s\n\n", branch.Name)
	}

	if short || porcelain {
		printShort(paths, headFiles, indexFiles, workFiles, untracked, ignored, short)
		return nil
	}

	printSectionStaged("new file", stagedAdded)
	printSectionStaged("modified", stagedModified)
	printSectionStaged("deleted", stagedDeleted)
	if len(stagedAdded)+len(stagedModified)+len(stagedDeleted) > 0 {
		fmt.Println()
	}

	printSection("modified", unstagedModified)
	printSection("deleted", unstagedDeleted)
	if len(unstagedModified)+len(unstagedDeleted) > 0 {
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		fmt.Println("  (use \"bvc add <file>...\" to include in what will be committed)")
		for _, p := range untracked {
			fmt.Printf("\t%s\n", rel(p))
		}
		fmt.Println()
	}

	if showIgnored && len(ignored) > 0 {
		fmt.Println("Ignored files:")
		for _, p := range ignored {
			fmt.Printf("    %s\n", rel(p))
		}
		fmt.Println()
	}

	if len(stagedAdded)+len(stagedModified)+len(stagedDeleted)+
		len(unstagedModified)+len(unstagedDeleted)+len(untracked)+len(ignored) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	}

	return nil
}

// helpers
func rel(p string) string {
	wd, _ := filepath.Abs(".")
	if r, err := filepath.Rel(wd, p); err == nil {
		return r
	}
	return p
}

func printSectionStaged(kind string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Println("Changes to be committed:")
	fmt.Println("  (use \"bvc restore --staged <file>...\" to unstage)")
	for _, p := range items {
		fmt.Printf("\t%s:   %s\n", kind, rel(p))
	}
}

func printSection(kind string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Println("Changes not staged for commit:")
	fmt.Println("  (use \"bvc add <file>...\" to update what will be committed)")
	for _, p := range items {
		fmt.Printf("\t%s:   %s\n", kind, rel(p))
	}
}

func printShort(
	paths []string,
	head, index, work map[string]file.Entry,
	untracked []string,
	ignored []string,
	short bool,
) {
	for _, p := range paths {
		h, inHead := head[p]
		i, inIndex := index[p]
		w, inWork := work[p]

		var x, y string

		if inIndex {
			if !inHead {
				x = "A"
			} else if len(i.Blocks) == 0 {
				x = "D"
			} else if !h.Equal(&i) {
				x = "M"
			}
		}

		if inWork && inIndex {
			if !i.Equal(&w) {
				y = "M"
			}
		} else if inWork && inHead && !inIndex {
			if !h.Equal(&w) {
				y = "M"
			}
		} else if inWork && !inHead && !inIndex {
			// do nothing; untracked will be handled in separate loop
		} else if !inWork && inHead && !inIndex {
			y = "D"
		}

		if x != "" || y != "" {
			line := fmt.Sprintf("%s%s %s", x, y, rel(p))
			if short {
				line = colorXY(line, x, y)
			}
			fmt.Println(line)
		}
	}

	for _, p := range untracked {
		line := fmt.Sprintf("?? %s", rel(p))
		if short {
			line = "\033[36m" + line + "\033[0m"
		}
		fmt.Println(line)
	}

	for _, p := range ignored {
		line := fmt.Sprintf("!! %s", rel(p))
		if short {
			line = "\033[90m" + line + "\033[0m"
		}
		fmt.Println(line)
	}
}

func colorXY(line, x, y string) string {
	if x == "A" || y == "A" {
		return "\033[32m" + line + "\033[0m"
	} else if x == "D" || y == "D" {
		return "\033[31m" + line + "\033[0m"
	} else if x == "M" || y == "M" {
		return "\033[33m" + line + "\033[0m"
	}
	return line
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
