package status

import (
	"flag"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/internal/repo/store/file"
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

type statusItem struct {
	Path     string
	Staged   string // "A", "M", "D"
	Unstaged string // "M", "D"
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

	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	branch, err := r.Meta.GetCurrentBranch()
	if err != nil {
		if !quiet {
			fmt.Println("No commits yet on current branch")
		}
		return nil
	}

	// head files
	headFiles := map[string]file.Entry{}
	if commitID, _ := r.Meta.GetLastCommitID(branch.Name); commitID != "" {
		fs, err := r.GetCommittedFileset(commitID)
		if err != nil {
			return err
		}
		for _, e := range fs.Files {
			headFiles[filepath.Clean(e.Path)] = e
		}
	}

	// work, staged, ignored filesets
	workFS, stagedFS, ignoredFS, err := r.Store.Snapshots.BuildAllRepositoryFilesets()
	if err != nil {
		return fmt.Errorf("scan working tree: %w", err)
	}

	workFiles := map[string]file.Entry{}
	for _, e := range workFS.Files {
		workFiles[filepath.Clean(e.Path)] = e
	}

	stagedFiles := map[string]file.Entry{}
	for _, e := range stagedFS.Files {
		stagedFiles[filepath.Clean(e.Path)] = e
	}

	ignoredFiles := map[string]file.Entry{}
	for _, e := range ignoredFS.Files {
		ignoredFiles[filepath.Clean(e.Path)] = e
	}

	// collect all unique paths
	allPaths := make(map[string]struct{})
	for k := range headFiles {
		allPaths[k] = struct{}{}
	}
	for k := range stagedFiles {
		allPaths[k] = struct{}{}
	}
	for k := range workFiles {
		allPaths[k] = struct{}{}
	}

	paths := make([]string, 0, len(allPaths))
	for p := range allPaths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var statusList []statusItem
	var untracked []string

	for _, p := range paths {
		h, inHead := headFiles[p]
		s, inStaged := stagedFiles[p]
		w, inWork := workFiles[p]

		var staged, unstaged string

		// determine staged status
		switch {
		case inStaged && !inHead:
			staged = "A"
		case inStaged && inHead && !h.Equal(&s):
			staged = "M"
		case inHead && !inStaged:
			staged = "D"
		}

		// determine unstaged status
		switch {
		case inWork && inStaged && !s.Equal(&w):
			unstaged = "M"
		case inWork && !inStaged && inHead && !h.Equal(&w):
			unstaged = "M"
		case !inWork && inHead && inStaged:
			unstaged = "D"
		}

		if staged != "" || unstaged != "" {
			statusList = append(statusList, statusItem{
				Path:     p,
				Staged:   staged,
				Unstaged: unstaged,
			})
			continue
		}

		// determine untracked
		if !inStaged && !inHead && inWork && untrackedMode != "no" {
			untracked = append(untracked, p)
		}
	}

	// collect ignored list
	var ignoredList []string
	if showIgnored {
		for _, e := range ignoredFS.Files {
			ignoredList = append(ignoredList, e.Path)
		}
		sort.Strings(ignoredList)
	}

	if quiet {
		return nil
	}

	if showBranch || (!short && !porcelain) {
		fmt.Printf("On branch %s\n\n", branch.Name)
	}

	// render status with colors if not porcelain
	if short || porcelain {
		printShortStatus(statusList, untracked, ignoredList, !porcelain)
	} else {
		printFullStatus(statusList, untracked, ignoredList, !porcelain)
	}

	// show clean tree message
	if len(statusList) == 0 && len(untracked) == 0 && len(ignoredList) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	}

	return nil
}

func printShortStatus(items []statusItem, untracked, ignored []string, color bool) {
	for _, it := range items {
		line := fmt.Sprintf("%s%s %s", it.Staged, it.Unstaged, rel(it.Path))
		if color {
			line = colorLine(it.Staged, it.Unstaged, line)
		}
		fmt.Println(line)
	}

	for _, u := range untracked {
		line := fmt.Sprintf("?? %s", rel(u))
		if color {
			line = "\033[31m" + line + "\033[0m" // red
		}
		fmt.Println(line)
	}

	for _, i := range ignored {
		line := fmt.Sprintf("!! %s", rel(i))
		if color {
			line = "\033[90m" + line + "\033[0m" // gray
		}
		fmt.Println(line)
	}
}

func printFullStatus(items []statusItem, untracked, ignored []string, color bool) {
	var staged, unstaged []statusItem
	for _, it := range items {
		if it.Staged != "" {
			staged = append(staged, it)
		}
		if it.Unstaged != "" {
			unstaged = append(unstaged, it)
		}
	}

	if len(staged) > 0 {
		fmt.Println("Changes to be committed:")
		fmt.Println("  (use \"bvc restore --staged <file>...\" to unstage)")
		for _, it := range staged {
			kindStr := kind(it.Staged)
			line := fmt.Sprintf("\t%-10s %s", kindStr+":", rel(it.Path))
			if color {
				line = colorLine(it.Staged, "", line)
			}
			fmt.Println(line)
		}
		fmt.Println()
	}

	if len(unstaged) > 0 {
		fmt.Println("Changes not staged for commit:")
		fmt.Println("  (use \"bvc add <file>...\" to update what will be committed)")
		for _, it := range unstaged {
			kindStr := kind(it.Unstaged)
			line := fmt.Sprintf("\t%-10s %s", kindStr+":", rel(it.Path))
			if color {
				line = colorLine(it.Unstaged, "", line)
			}
			fmt.Println(line)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		fmt.Println("  (use \"bvc add <file>...\" to include in what will be committed)")
		for _, u := range untracked {
			line := fmt.Sprintf("\t%s", rel(u))
			if color {
				line = "\033[31m" + line + "\033[0m" // red
			}
			fmt.Println(line)
		}
		fmt.Println()
	}

	if len(ignored) > 0 {
		fmt.Println("Ignored files:")
		fmt.Println("  (use \"bvc add -f <file>...\" to include in what will be committed)")
		for _, i := range ignored {
			line := fmt.Sprintf("\t%s", rel(i))
			if color {
				line = "\033[90m" + line + "\033[0m" // gray
			}
			fmt.Println(line)
		}
		fmt.Println()
	}
}

func colorLine(staged, unstaged, line string) string {
	switch {
	case staged == "A" || unstaged == "A":
		return "\033[32m" + line + "\033[0m" // green
	case staged == "M" || unstaged == "M":
		return "\033[33m" + line + "\033[0m" // yellow
	case staged == "D" || unstaged == "D":
		return "\033[31m" + line + "\033[0m" // red
	default:
		return line
	}
}

func kind(x string) string {
	switch x {
	case "A":
		return "new file"
	case "M":
		return "modified"
	case "D":
		return "deleted"
	default:
		return x
	}
}

func rel(p string) string {
	wd, _ := filepath.Abs(".")
	if r, err := filepath.Rel(wd, p); err == nil {
		return r
	}
	return p
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
