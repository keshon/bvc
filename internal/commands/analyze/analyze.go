package analyze

import (
	"app/internal/cli"
	"app/internal/middleware"
	"app/internal/storage/snapshot"
	"fmt"
	"sort"
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
	// Загружаем все filesets
	fss, err := snapshot.LoadFilesets()
	if err != nil {
		return err
	}
	if len(fss) == 0 {
		fmt.Println("No filesets found.")
		return nil
	}

	// Словари для подсчёта блоков и файлов
	blockCounts := map[string]int{}                // hash -> количество вхождений
	fileBlocks := map[string]map[string]struct{}{} // file path -> set of block hashes

	for _, fs := range fss {
		for _, file := range fs.Files {
			if _, ok := fileBlocks[file.Path]; !ok {
				fileBlocks[file.Path] = map[string]struct{}{}
			}
			for _, blk := range file.Blocks {
				fileBlocks[file.Path][blk.Hash] = struct{}{}
				blockCounts[blk.Hash]++
			}
		}
	}

	// Общие статистики
	totalBlocks := len(blockCounts)
	sharedBlocks := 0
	for _, count := range blockCounts {
		if count > 1 {
			sharedBlocks++
		}
	}

	// Статистика по файлам
	type FileStat struct {
		Path         string
		TotalBlocks  int
		SharedBlocks int
		ReuseRatio   float64
	}
	var stats []*FileStat
	for path, blocks := range fileBlocks {
		fs := &FileStat{Path: path}
		for h := range blocks {
			fs.TotalBlocks++
			if blockCounts[h] > 1 {
				fs.SharedBlocks++
			}
		}
		if fs.TotalBlocks > 0 {
			fs.ReuseRatio = float64(fs.SharedBlocks) / float64(fs.TotalBlocks) * 100
		}
		stats = append(stats, fs)
	}

	// Сортировка по повторяемости
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].ReuseRatio > stats[j].ReuseRatio
	})

	// Вывод
	fmt.Println("\033[90m────────────────────────────────────\033[0m")
	fmt.Printf("\033[90mTotal blocks:\033[0m  %d\n", totalBlocks)
	fmt.Printf("\033[90mUnique blocks:\033[0m %d\n", totalBlocks-sharedBlocks)
	fmt.Printf("\033[90mShared blocks:\033[0m %d\n", sharedBlocks)
	fmt.Printf("\033[90mReuse ratio:\033[0m   %.1f%%\n", float64(sharedBlocks)/float64(totalBlocks)*100)
	fmt.Println("\033[90m────────────────────────────────────\033[0m")
	fmt.Println("Top reused files:")
	for i := 0; i < len(stats) && i < 5; i++ {
		f := stats[i]
		fmt.Printf("  %-40s %5.1f%% (%d blocks)\n", f.Path, f.ReuseRatio, f.TotalBlocks)
	}

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
