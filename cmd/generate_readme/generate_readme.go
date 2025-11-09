package main

import (
	"app/internal/command"
	"fmt"
	"os"
	"sort"
	"text/template"

	_ "app/internal/command/add"
	_ "app/internal/command/block"
	_ "app/internal/command/branch"
	_ "app/internal/command/checkout"
	_ "app/internal/command/cherry-pick"
	_ "app/internal/command/commit"
	_ "app/internal/command/help"
	_ "app/internal/command/init"
	_ "app/internal/command/log"
	_ "app/internal/command/merge"
	_ "app/internal/command/reset"
	_ "app/internal/command/status"
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
			"### %s\n```\n%s\n%s\n```\n\n",
			cmd.Name(),
			cmd.Usage(),
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
