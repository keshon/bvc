package main

import (
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/keshon/bvc/internal/command"

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

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	sections := ""
	for _, cmd := range commands {
		sections += fmt.Sprintf(
			"### bvc %s\n```\n%s\n```\n\n",
			cmd.Name(),
			cmd.Help(),
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
