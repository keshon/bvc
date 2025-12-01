package main

import (
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/keshon/bvc/internal/command"

	// register all commands
	_ "github.com/keshon/bvc/internal/command/add"
	_ "github.com/keshon/bvc/internal/command/block"
	_ "github.com/keshon/bvc/internal/command/branch"
	_ "github.com/keshon/bvc/internal/command/checkout"
	_ "github.com/keshon/bvc/internal/command/cherry-pick"
	_ "github.com/keshon/bvc/internal/command/commit"
	_ "github.com/keshon/bvc/internal/command/help"
	_ "github.com/keshon/bvc/internal/command/init"
	_ "github.com/keshon/bvc/internal/command/log"
	_ "github.com/keshon/bvc/internal/command/merge"
	_ "github.com/keshon/bvc/internal/command/reset"
	_ "github.com/keshon/bvc/internal/command/status"
)

func main() {
	tplBytes, err := os.ReadFile("README.md.tmpl")
	if err != nil {
		fmt.Printf("Failed to read template: %v\n", err)
		os.Exit(1)
	}

	tpl, err := template.New("readme").Parse(string(tplBytes))
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		os.Exit(1)
	}

	commands := command.AllCommands()

	// prepare a slice to hold all leaf command entries
	type CmdEntry struct {
		Path string
		Cmd  command.Command
	}
	var entries []CmdEntry

	// walk recursively to collect all leaf commands
	var walk func(prefix string, cmd command.Command)
	walk = func(prefix string, cmd command.Command) {
		path := cmd.Name()
		if prefix != "" {
			path = prefix + " " + cmd.Name()
		}

		if len(cmd.Subcommands()) == 0 {
			entries = append(entries, CmdEntry{Path: path, Cmd: cmd})
		}

		for _, sc := range cmd.Subcommands() {
			walk(path, sc)
		}
	}

	for _, cmd := range commands {
		walk("", cmd)
	}

	// sort entries alphabetically by full path
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	// generate sections
	sections := ""
	for _, e := range entries {
		sections += fmt.Sprintf(
			"### bvc %s\n```\n%s\n```\n\n",
			e.Path,
			e.Cmd.Help(),
		)
	}

	data := map[string]string{
		"CommandSections": sections,
	}

	outFile, err := os.Create("README.md")
	if err != nil {
		fmt.Printf("Failed to create README.md: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	if err := tpl.Execute(outFile, data); err != nil {
		fmt.Printf("Failed to render template: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("README.md generated successfully")
}
