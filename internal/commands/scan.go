package commands

import (
	"fmt"
	"time"

	"app/internal/cli"
	"app/internal/storage/block"
	"app/internal/verify"
)

// ScanCommand checks repository block integrity and file consistency
type ScanCommand struct{}

// Canonical name
func (c *ScanCommand) Name() string { return "scan" }

// Usage string
func (c *ScanCommand) Usage() string { return "scan" }

// Short description
func (c *ScanCommand) Description() string {
	return "Verify repository blocks and file integrity"
}

// Detailed description
func (c *ScanCommand) DetailedDescription() string {
	return "Scan repository blocks, verify file integrity, and detect leftover temporary files."
}

// Aliases
func (c *ScanCommand) Aliases() []string { return []string{"chk"} }

// Shortcut
func (c *ScanCommand) Short() string { return "I" }

// Run executes the scan
func (c *ScanCommand) Run(ctx *cli.Context) error {
	// Cleanup temporary files first
	if err := block.CleanupTmp(); err != nil {
		fmt.Printf("Warning: tmp file cleanup failed: %v\n", err)
	}

	// Stream block verification
	out, errCh := verify.ScanRepositoryBlocksStream()

	// Print legend
	fmt.Print("\033[90mLegend:\033[0m \033[32m█\033[0m OK   \033[31m█\033[0m Missing   \033[33m█\033[0m Damaged\n\n")

	start := time.Now()
	count, okCount, missingCount, damagedCount := 0, 0, 0, 0

	// Collect results
	results := []block.BlockCheck{}

	for out != nil || errCh != nil {
		select {
		case bc, ok := <-out:
			if !ok {
				out = nil
				continue
			}
			results = append(results, bc)

			// Print colored status bar
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

	// Summary
	fmt.Printf("\nScan complete in %s.\n", time.Since(start).Truncate(time.Millisecond))
	fmt.Printf("Blocks OK: \033[32m%d\033[0m   Missing: \033[31m%d\033[0m   Damaged: \033[33m%d\033[0m\n",
		okCount, missingCount, damagedCount)

	if missingCount+damagedCount > 0 {
		fmt.Println("\nSome blocks may need repair. Run `bvc repair`.")
	}

	return nil
}

// Register command
func init() {
	cli.RegisterCommand(&ScanCommand{})
}
