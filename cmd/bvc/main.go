package main

import (
	"fmt"
	"os"

	"bvc/internal/commands/help"
	"bvc/internal/commands/initrepo"
	"bvc/internal/commands/prune"
	"bvc/internal/commands/snapshot"
	"bvc/internal/commands/status"
	"bvc/internal/commands/stream"
	"bvc/internal/commands/sync"
	"bvc/pkg/command"
)

func main() {
	reg := command.NewRegistry()

	reg.Register(&help.HelpCmd{Reg: reg})
	reg.Register(&initrepo.InitCmd{})
	reg.Register(&prune.PruneCmd{})
	reg.Register(&snapshot.SnapshotCmd{})
	reg.Register(&status.StatusCmd{})
	reg.Register(&stream.StreamCmd{})
	reg.Register(&sync.SyncCmd{})

	if err := reg.Dispatch(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
