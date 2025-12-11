package main

import (
	"fmt"
	"os"

	"github.com/keshon/bvc/internal/commands/check"
	"github.com/keshon/bvc/internal/commands/help"
	"github.com/keshon/bvc/internal/commands/initrepo"
	"github.com/keshon/bvc/internal/commands/prune"
	"github.com/keshon/bvc/internal/commands/snapshot"
	"github.com/keshon/bvc/internal/commands/status"
	"github.com/keshon/bvc/internal/commands/stream"
	"github.com/keshon/bvc/internal/commands/sync"

	"github.com/keshon/bvc/pkg/command"
)

func main() {
	reg := command.DefaultRegistry

	reg.Register(&help.HelpCmd{})
	reg.Register(&initrepo.InitCmd{})
	reg.Register(&prune.PruneCmd{})
	reg.Register(&snapshot.SnapshotCmd{})
	reg.Register(&status.StatusCmd{})
	reg.Register(&stream.StreamCmd{})
	reg.Register(&sync.SyncCmd{})
	reg.Register(&check.CheckCmd{})

	if err := reg.DispatchCLI(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
