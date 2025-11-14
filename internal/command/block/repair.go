package block

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/fs"

	"app/internal/repo"
	"app/internal/repo/store/block"
	"app/internal/repotools"
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"time"
)

type RepairCommand struct{}

func (c *RepairCommand) Name() string  { return "repair" }
func (c *RepairCommand) Brief() string { return "Repair missing or damaged repository blocks" }
func (c *RepairCommand) Usage() string { return "block repair" }
func (c *RepairCommand) Help() string {
	return "Repair any missing or damaged blocks automatically."
}
func (c *RepairCommand) Aliases() []string              { return []string{"verify-repair"} }
func (c *RepairCommand) Subcommands() []command.Command { return nil }
func (c *RepairCommand) Flags(fs *flag.FlagSet)         {}

func (c *RepairCommand) Run(ctx *command.Context) error {
	fs := fs.NewOSFS()
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	out, errCh := repotools.VerifyBlocksStream(r.Meta, r.Config, true)

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
		_ = fs.Remove(targetPath)

		fixed := false

		for _, currFile := range bc.Files {
			entry, err := r.Store.Files.BuildEntry(currFile)
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
					_ = fs.Remove(targetPath)
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
