package commands

import (
	"app/internal/cli"
	"app/internal/storage"
	"app/internal/verify"
	"fmt"
	"time"
)

type ScanCommand struct{}

func (c *ScanCommand) Name() string        { return "scan" }
func (c *ScanCommand) Description() string { return "Verify repository blocks and file integrity" }
func (c *ScanCommand) Usage() string       { return "scan" }
func (c *ScanCommand) DetailedDescription() string {
	return "Scan repository blocks and verify file integrity"
}

func (c *ScanCommand) Run(ctx *cli.Context) error {
	results := []storage.BlockCheck{}

	out, errCh := verify.ScanRepositoryBlocksStream()

	fmt.Print("\033[90mLegend:\033[0m \033[32m█\033[0m OK   \033[31m█\033[0m Missing   \033[33m█\033[0m Damaged\n\n")

	start := time.Now()
	lineWidth := 100
	count := 0
	okCount, missingCount, damagedCount := 0, 0, 0

	for out != nil || errCh != nil {
		select {
		case bc, ok := <-out:
			if !ok {
				out = nil
				continue
			}
			results = append(results, bc)

			switch bc.Status {
			case storage.BlockOK:
				fmt.Print("\033[32m█\033[0m")
				okCount++
			case storage.BlockMissing:
				fmt.Print("\033[31m█\033[0m")
				missingCount++
			case storage.BlockDamaged:
				fmt.Print("\033[33m█\033[0m")
				damagedCount++
			}

			count++
			if count%lineWidth == 0 {
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

	if count%lineWidth != 0 {
		fmt.Printf("  %d\n", count)
	}

	fmt.Printf("\nScan complete in %s.\n", time.Since(start).Truncate(time.Millisecond))
	fmt.Printf("Blocks OK: \033[32m%d\033[0m   Missing: \033[31m%d\033[0m   Damaged: \033[33m%d\033[0m\n",
		okCount, missingCount, damagedCount)

	if missingCount > 0 {
		fmt.Println("\nMissing blocks:")
		for _, bc := range results {
			if bc.Status == storage.BlockMissing {
				fmt.Printf("\033[31m%s\033[0m  files: %v  branches: %v\n",
					bc.Hash, bc.Files, bc.Branches)
			}
		}
	}

	if damagedCount > 0 {
		fmt.Println("\nDamaged blocks:")
		for _, bc := range results {
			if bc.Status == storage.BlockDamaged {
				fmt.Printf("\033[33m%s\033[0m  files: %v  branches: %v\n",
					bc.Hash, bc.Files, bc.Branches)
			}
		}
	}

	return nil
}

func init() { cli.RegisterCommand(&ScanCommand{}) }
