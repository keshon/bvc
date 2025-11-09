package block

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/repo"
	"app/internal/repo/store/block"
	"app/internal/repotools"
	"flag"
	"fmt"
	"time"
)

type ScanCommand struct{}

func (c *ScanCommand) Name() string  { return "scan" }
func (c *ScanCommand) Brief() string { return "Scan repository blocks for integrity issues" }
func (c *ScanCommand) Usage() string { return "block scan" }
func (c *ScanCommand) Help() string {
	return "Scan all repository blocks and report missing or damaged ones."
}
func (c *ScanCommand) Aliases() []string              { return []string{"verify"} }
func (c *ScanCommand) Subcommands() []command.Command { return nil }
func (c *ScanCommand) Flags(fs *flag.FlagSet)         {}

func (c *ScanCommand) Run(ctx *command.Context) error {
	r, err := repo.NewRepositoryByPath(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	out, errCh := repotools.VerifyBlocksStream(r.Meta, r.Config, true)

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
