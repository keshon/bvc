package status

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/middleware"
	"github.com/keshon/bvc/internal/repo"
)

type Command struct {
	short          bool
	porcelain      bool
	branch         bool
	untrackedFiles string
	ignored        bool
	quiet          bool
}

func (c *Command) Name() string      { return "status" }
func (c *Command) Aliases() []string { return []string{"st"} }
func (c *Command) Brief() string     { return "Show working tree and index status" }
func (c *Command) Usage() string     { return "status [options]" }

func (c *Command) Help() string {
	return `Show the working tree status.

Options:
  -s, --short                    Show short summary (XY path)
      --porcelain                Machine-readable short output
  -b, --branch                   Show branch info
  -u, --untracked-files=<mode>   Show untracked files: no, normal, all (default: normal)
      --ignored                  Show ignored files
  -q, --quiet                    Suppress normal output

Usage:
  bvc status [options]

Examples:
  bvc status
  bvc status -s
  bvc status --branch
`
}

func (c *Command) Subcommands() []command.Command { return nil }

func (c *Command) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.short, "short", false, "show short summary")
	fs.BoolVar(&c.short, "s", false, "alias for --short")

	fs.BoolVar(&c.porcelain, "porcelain", false, "machine-readable short output")

	fs.BoolVar(&c.branch, "branch", false, "")
	fs.BoolVar(&c.branch, "b", false, "alias for --branch")

	fs.StringVar(&c.untrackedFiles, "untracked-files", "normal", "")
	fs.StringVar(&c.untrackedFiles, "u", "normal", "alias for --untracked-files")

	fs.BoolVar(&c.ignored, "ignored", false, "")

	fs.BoolVar(&c.quiet, "quiet", false, "")
	fs.BoolVar(&c.quiet, "q", false, "alias for --quiet")
}

func (c *Command) Run(ctx *command.Context) error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return err
	}

	st, err := r.Status(repo.StatusOptions{
		UntrackedMode: c.untrackedFiles,
		ShowIgnored:   c.ignored,
	})
	if err != nil {
		return err
	}

	if c.quiet {
		return nil
	}

	if c.branch || (!c.short && !c.porcelain) {
		fmt.Printf("On branch %s\n\n", st.Branch)
	}

	if c.short || c.porcelain {
		printShortStatus(st.Items, st.Untracked, st.Ignored, !c.porcelain)
	} else {
		printFullStatus(st.Items, st.Untracked, st.Ignored, !c.porcelain)
	}

	if st.IsClean {
		fmt.Println("nothing to commit, working tree clean")
	}

	return nil
}

func printShortStatus(items []repo.FileStatus, untracked, ignored []string, color bool) {
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

func printFullStatus(items []repo.FileStatus, untracked, ignored []string, color bool) {
	var staged, unstaged []repo.FileStatus
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
