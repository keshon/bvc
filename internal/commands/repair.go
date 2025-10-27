package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/verify"

	"github.com/zeebo/xxh3"
)

// RepairCommand repairs missing or damaged repository blocks.
type RepairCommand struct{}

// Canonical name
func (c *RepairCommand) Name() string { return "repair" }

// Usage string
func (c *RepairCommand) Usage() string { return "repair" }

// Short description
func (c *RepairCommand) Description() string {
	return "Repair missing or damaged repository blocks"
}

// Detailed description
func (c *RepairCommand) DetailedDescription() string {
	return "Repair missing or damaged blocks from repository files. Scans blocks and attempts to restore from known files."
}

// Aliases
func (c *RepairCommand) Aliases() []string { return []string{"fix"} }

// Shortcut
func (c *RepairCommand) Short() string { return "R" }

// Run executes the repair process
func (c *RepairCommand) Run(ctx *cli.Context) error {
	// Stream blocks to identify missing or damaged ones
	out, errCh := verify.ScanRepositoryBlocksStream()

	fmt.Print("\033[90mLegend:\033[0m \033[32m█\033[0m OK   \033[31m█\033[0m Failed\n\n")

	var toFix []block.BlockCheck

	// Stream results safely to avoid deadlocks
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

	var fixedList []block.BlockCheck
	var failedList []block.BlockCheck

	for _, bc := range toFix {
		targetPath := filepath.Join(config.ObjectsDir, bc.Hash+".bin")

		// Remove any damaged or stale block file
		_ = os.Remove(targetPath)
		fixed := false

		// Try to rebuild block from known files
		for _, currFile := range bc.Files {
			entry, err := file.Build(currFile)
			if err != nil {
				continue
			}

			for _, b := range entry.Blocks {
				if b.Hash != bc.Hash {
					continue
				}

				// Store the block
				if err := block.Store(entry.Path, []block.BlockRef{b}); err != nil {
					continue
				}

				// Verify integrity
				status, _ := block.Verify(b.Hash)
				if status == block.OK {
					fixed = true
					repaired++
					break
				} else {
					_ = os.Remove(targetPath) // delete invalid block
				}
			}
			if fixed {
				break
			}
		}

		if fixed {
			fmt.Print("\033[32m█\033[0m") // success
			fixedList = append(fixedList, bc)
		} else {
			fmt.Print("\033[31m█\033[0m") // failure
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

	// Summary listing
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

// verifyBlockHash ensures the on-disk block matches expected hash.
func verifyBlockHash(path, expected string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	sum := fmt.Sprintf("%x", xxh3.Hash128(data).Bytes())
	return sum == expected, nil
}

// verifyRepairedBlocks re-checks integrity after repair and lists failures.
func verifyRepairedBlocks(toFix []block.BlockCheck) int {
	fmt.Println("\nVerifying repaired blocks...")
	failed := 0

	for _, bc := range toFix {
		path := filepath.Join(config.ObjectsDir, bc.Hash+".bin")
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

// Register command
func init() {
	cli.RegisterCommand(&RepairCommand{})
}
