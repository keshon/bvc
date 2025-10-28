package commands

import (
	"app/internal/cli"
	"app/internal/repo"
	"fmt"
	"sort"
	"strings"
)

type AnalyzeCommand struct{}

func (c *AnalyzeCommand) Name() string  { return "analyze" }
func (c *AnalyzeCommand) Usage() string { return "analyze [--sort reuse|unique|size] [--json]" }
func (c *AnalyzeCommand) Description() string {
	return "Analyze block reuse across the entire repository (all snapshots and branches)"
}

func (c *AnalyzeCommand) DetailedDescription() string {
	return "Analyze block reuse across the entire repository"
}

func (c *AnalyzeCommand) Aliases() []string { return []string{"a"} }

func (c *AnalyzeCommand) Short() string { return "a" }

func (c *AnalyzeCommand) Run(ctx *cli.Context) error {
	sortMode := "reuse"
	for k, v := range ctx.Flags {
		switch strings.ToLower(k) {
		case "sort":
			sortMode = strings.ToLower(v)
		case "json":
			// future: output json
		}
	}

	fmt.Println("Scanning repository...")
	all, err := repo.CollectAllBlocks()
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

	var totalBlocks, totalSize, sharedBlocks int

	for _, info := range all {
		totalBlocks++
		totalSize += int(info.Size)
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

	// compute ratios
	for _, f := range fileStats {
		if f.TotalBlocks > 0 {
			f.ReuseRatio = float64(f.SharedBlocks) / float64(f.TotalBlocks) * 100
		}
	}

	fmt.Println("────────────────────────────────────")
	fmt.Printf("Total blocks:  %d\n", totalBlocks)
	fmt.Printf("Unique blocks: %d\n", totalBlocks-sharedBlocks)
	fmt.Printf("Shared blocks: %d\n", sharedBlocks)
	fmt.Printf("Reuse ratio:   %.1f%%\n", float64(sharedBlocks)/float64(totalBlocks)*100)
	fmt.Printf("────────────────────────────────────\n")

	// sort by chosen metric
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

	fmt.Println("Top reused files:")
	for i := 0; i < len(stats) && i < 5; i++ {
		f := stats[i]
		fmt.Printf("  %-40s %5.1f%% (%d blocks)\n", f.Path, f.ReuseRatio, f.TotalBlocks)
	}

	fmt.Println("────────────────────────────────────")
	return nil
}

func init() {
	cli.RegisterCommand(&AnalyzeCommand{})
}
