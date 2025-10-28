package analyze

import (
	"app/internal/cli"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
	"sort"
	"strings"
)

type Command struct{}

func (c *Command) Name() string  { return "analyze" }
func (c *Command) Usage() string { return "analyze [--sort reuse|unique|size]" }
func (c *Command) Description() string {
	return "Analyze block reuse across the entire repository (all snapshots and branches)"
}

func (c *Command) DetailedDescription() string {
	return "Analyze block reuse across the entire repository"
}

func (c *Command) Aliases() []string { return []string{"a"} }

func (c *Command) Short() string { return "a" }

func (c *Command) Run(ctx *cli.Context) error {
	sortMode := "reuse"

	// Parse arguments
	for i := 0; i < len(ctx.Args); i++ {
		switch strings.ToLower(ctx.Args[i]) {
		case "--sort":
			if i+1 < len(ctx.Args) {
				sortMode = strings.ToLower(ctx.Args[i+1])
				i++
			}
		}
	}

	fmt.Println("Scanning repository...")
	all, err := repo.ListAllBlocks(true) // allHistory = true
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Println("No blocks found.")
		return nil
	}

	type FileStat struct {
		Path         string
		TotalBlocks  int
		SharedBlocks int
		ReuseRatio   float64
	}

	fileStats := map[string]*FileStat{}
	var totalBlocks, sharedBlocks int

	// Collect per-file stats
	for _, info := range all {
		totalBlocks++
		if len(info.Files) > 1 {
			sharedBlocks++
		}

		for path := range info.Files {
			fs, ok := fileStats[path]
			if !ok {
				fs = &FileStat{Path: path}
				fileStats[path] = fs
			}
			fs.TotalBlocks++
			if len(info.Files) > 1 {
				fs.SharedBlocks++
			}
		}
	}

	// Compute reuse ratio per file
	for _, f := range fileStats {
		if f.TotalBlocks > 0 {
			f.ReuseRatio = float64(f.SharedBlocks) / float64(f.TotalBlocks) * 100
		}
	}

	// Print summary
	fmt.Println("────────────────────────────────────")
	fmt.Printf("Total blocks:  %d\n", totalBlocks)
	fmt.Printf("Unique blocks: %d\n", totalBlocks-sharedBlocks)
	fmt.Printf("Shared blocks: %d\n", sharedBlocks)
	fmt.Printf("Reuse ratio:   %.1f%%\n", float64(sharedBlocks)/float64(totalBlocks)*100)
	fmt.Println("────────────────────────────────────")

	// Sort files according to chosen metric
	stats := make([]*FileStat, 0, len(fileStats))
	for _, f := range fileStats {
		stats = append(stats, f)
	}

	switch sortMode {
	case "unique":
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].ReuseRatio < stats[j].ReuseRatio
		})
	case "size":
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].TotalBlocks > stats[j].TotalBlocks
		})
	default: // reuse
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].ReuseRatio > stats[j].ReuseRatio
		})
	}

	// Show top reused files
	fmt.Println("Top reused files:")
	for i := 0; i < len(stats) && i < 5; i++ {
		f := stats[i]
		fmt.Printf("  %-60s %5.1f%% (%d blocks)\n", f.Path, f.ReuseRatio, f.TotalBlocks)
	}

	fmt.Println("────────────────────────────────────")
	return nil
}

func init() {
	cli.RegisterCommand(
		cli.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
