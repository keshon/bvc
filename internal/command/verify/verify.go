package veirfy

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/fsio"
	"app/internal/middleware"
	"app/internal/repo"
	"app/internal/repo/store/block"
	"app/internal/repotools"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/zeebo/xxh3"
)

type Command struct{}

func (c *Command) Name() string      { return "verify" }
func (c *Command) Short() string     { return "V" }
func (c *Command) Aliases() []string { return []string{"scan", "check"} }
func (c *Command) Usage() string     { return "verify [--repair|--auto]" }
func (c *Command) Brief() string     { return "Verify or repair repository integrity" }
func (c *Command) Help() string {
	return `Verify repository blocks and file integrity.

Usage:
  verify           - Scan all blocks and report missing/damaged ones.
  verify --repair  - Attempt to repair any missing or damaged blocks automatically.
`
}

func (c *Command) Run(ctx *command.Context) error {
	for _, arg := range ctx.Args {
		if arg == "--repair" || arg == "-R" {
			return repair()
		}
	}

	return scan()
}

func scan() error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	out, errCh := repotools.VerifyBlocksStream(r, r.Config, true)

	fmt.Print("\033[90mLegend:\033[0m \033[32m█\033[0m OK   \033[31m█\033[0m Missing   \033[33m█\033[0m Damaged\n\n")

	start := time.Now()
	count, okCount, missingCount, damagedCount := 0, 0, 0, 0

	for out != nil || errCh != nil {
		select {
		case bc, ok := <-out:
			if !ok {
				out = nil
				continue
			}
			switch bc.Status {
			case block.OK:
				fmt.Print("\033[32m█\033[0m")
				okCount++
			case block.Missing:
				fmt.Print("\033[31m█\033[0m")
				missingCount++
			case block.Damaged:
				fmt.Print("\033[33m█\033[0m")
				damagedCount++
			}
			count++
			if count%100 == 0 {
				fmt.Printf("  %d\n", count)
			}

		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if err != nil {
				return err
			}
		}
	}

	if count%100 != 0 {
		fmt.Printf("  %d\n", count)
	}

	fmt.Printf("\nScan complete in %s.\n", time.Since(start).Truncate(time.Millisecond))
	fmt.Printf("Blocks OK: \033[32m%d\033[0m   Missing: \033[31m%d\033[0m   Damaged: \033[33m%d\033[0m\n",
		okCount, missingCount, damagedCount)

	if missingCount+damagedCount > 0 {
		fmt.Println("\nSome blocks may need repair. Run `bvc verify --repair`.")
	}

	return nil
}

func repair() error {
	// Open the repository context
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	out, errCh := repotools.VerifyBlocksStream(r, r.Config, true)

	fmt.Print("\033[90mLegend:\033[0m \033[32m█\033[0m OK   \033[31m█\033[0m Failed\n\n")

	var toFix []block.BlockCheck

	for out != nil || errCh != nil {
		select {
		case bc, ok := <-out:
			if !ok {
				out = nil
				continue
			}
			if bc.Status != block.OK {
				toFix = append(toFix, bc)
			}
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if err != nil {
				return err
			}
		}
	}

	if len(toFix) == 0 {
		fmt.Println("No missing or damaged blocks found. Nothing to repair.")
		return nil
	}

	fmt.Printf("Attempting to repair %d blocks...\n", len(toFix))
	start := time.Now()
	lineWidth := 50
	count := 0
	repaired := 0

	var fixedList, failedList []block.BlockCheck

	for _, bc := range toFix {
		targetPath := filepath.Join(r.Config.ObjectsDir(), bc.Hash+".bin")
		_ = fsio.Remove(targetPath)

		fixed := false

		for _, currFile := range bc.Files {
			entry, err := r.Store.Files.CreateEntry(currFile)
			if err != nil {
				continue
			}

			for _, b := range entry.Blocks {
				if b.Hash != bc.Hash {
					continue
				}
				if err := r.Store.Blocks.Write(entry.Path, []block.BlockRef{b}); err != nil {
					continue
				}
				status, _ := r.Store.Blocks.VerifyBlock(b.Hash)
				if status == block.OK {
					fixed = true
					repaired++
					break
				} else {
					_ = fsio.Remove(targetPath)
				}
			}
			if fixed {
				break
			}
		}

		if fixed {
			fmt.Print("\033[32m█\033[0m")
			fixedList = append(fixedList, bc)
		} else {
			fmt.Print("\033[31m█\033[0m")
			failedList = append(failedList, bc)
		}

		count++
		if count%lineWidth == 0 {
			fmt.Printf("  %d\n", count)
		}
	}

	if count%lineWidth != 0 {
		fmt.Printf("  %d\n", count)
	}

	fmt.Printf("\nRepair complete in %s.\n", time.Since(start).Truncate(time.Millisecond))
	fmt.Printf("Blocks repaired: \033[32m%d\033[0m / %d\n", repaired, len(toFix))

	// Final verification pass
	failed := verifyRepairedBlocks(toFix)

	if len(fixedList) > 0 {
		fmt.Println("\nRepaired blocks:")
		for _, bc := range fixedList {
			files := append([]string{}, bc.Files...)
			sort.Strings(files)
			fmt.Printf("\033[32m%s\033[0m  files: %v  branches: %v\n", bc.Hash, files, bc.Branches)
		}
	}

	if len(failedList) > 0 {
		fmt.Println("\nUnrepaired blocks:")
		for _, bc := range failedList {
			files := append([]string{}, bc.Files...)
			sort.Strings(files)
			fmt.Printf("\033[31m%s\033[0m  files: %v  branches: %v\n", bc.Hash, files, bc.Branches)
		}
	}

	if failed > 0 {
		fmt.Printf("\n\033[31m%d blocks remain corrupted or unrepaired.\033[0m\n", failed)
	} else {
		fmt.Println("\033[32mAll repaired blocks verified successfully.\033[0m")
	}

	return nil
}

// verifyRepairedBlocks re-checks integrity after repair
func verifyRepairedBlocks(toFix []block.BlockCheck) int {
	fmt.Println("\nVerifying repaired blocks...")
	failed := 0

	cfg := config.NewRepoConfig(config.ResolveRepoRoot())

	for _, bc := range toFix {
		path := filepath.Join(cfg.ObjectsDir(), bc.Hash+".bin")
		ok, _ := verifyBlockHash(path, bc.Hash)
		if !ok {
			failed++
			files := append([]string{}, bc.Files...)
			sort.Strings(files)
			fmt.Printf("\033[31m%s\033[0m  files: %v  branches: %v\n",
				bc.Hash, files, bc.Branches)
		}
	}
	return failed
}

// verifyBlockHash checks block hash consistency
func verifyBlockHash(path, expected string) (bool, error) {
	data, err := fsio.ReadFile(path)
	if err != nil {
		return false, err
	}
	sum := fmt.Sprintf("%x", xxh3.Hash128(data).Bytes())
	return sum == expected, nil
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
