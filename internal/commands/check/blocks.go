package check

import (
	"bvc/internal/repo"
	"bvc/pkg/command"
	"bvc/pkg/util"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
)

type BlocksCmd struct {
	repair bool
}

func (c *BlocksCmd) Name() string  { return "blocks" }
func (c *BlocksCmd) Brief() string { return "Verify integrity of stored blocks" }
func (c *BlocksCmd) Help() string  { return "Verify integrity of stored blocks" }
func (c *BlocksCmd) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.repair, "repair", false, "Attempt to repair missing blocks (not possible for blocks, just report)")
}
func (c *BlocksCmd) SubCommands() []command.Command { return nil }

func (c *BlocksCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	fmt.Println("Checking blocks integrity...")

	type fileBlocks struct {
		path   string
		blocks []string
	}

	// Map each block hash to all possible files that can generate it
	blockToFiles := map[string][]fileBlocks{}
	ids, err := r.Snaps.List()
	if err != nil {
		return err
	}

	for _, id := range ids {
		snap, err := r.Snaps.Load(id)
		if err != nil {
			continue
		}
		for rel, blocks := range snap.Files {
			absPath := r.Root + "/" + rel
			for i, h := range blocks {
				blockToFiles[h] = append(blockToFiles[h], fileBlocks{
					path:   absPath,
					blocks: blocks[i:], // remaining blocks in file
				})
			}
		}
	}

	var mu sync.Mutex
	missing := []string{}

	// First pass: detect missing/corrupted blocks
	hashes := make([]string, 0, len(blockToFiles))
	for h := range blockToFiles {
		hashes = append(hashes, h)
	}

	err = util.Parallel(hashes, 8, func(ctx context.Context, h string) error {
		if err := r.Blocks.Get(h, io.Discard, true); err != nil {
			mu.Lock()
			missing = append(missing, h)
			mu.Unlock()
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(missing) == 0 {
		fmt.Println("All blocks are intact.")
		return nil
	}

	fmt.Printf("Missing or corrupted blocks: %d\n", len(missing))
	for _, h := range missing {
		fmt.Println("  ", h)
	}

	if !c.repair {
		return nil
	}

	fmt.Println("Attempting repair using workspace files...")

	// Map from file path -> blocks that need to be recalculated from it
	fileMap := map[string]map[string]struct{}{}
	for _, h := range missing {
		for _, fb := range blockToFiles[h] {
			if _, ok := fileMap[fb.path]; !ok {
				fileMap[fb.path] = map[string]struct{}{}
			}
			fileMap[fb.path][h] = struct{}{}
		}
	}

	// Process each file once
	for path, missingSet := range fileMap {
		f, err := os.Open(path)
		if err != nil {
			fmt.Println("Failed to open file:", path, err)
			continue
		}

		buf := make([]byte, 4*1024*1024) // BlockSize
		for {
			n, err := f.Read(buf)
			if n > 0 {
				newHash, err := r.Blocks.PutIfAbsent(bytes.NewReader(buf[:n]))
				if err != nil {
					fmt.Println("Failed to store block:", err)
					break
				}
				// Remove restored block from missing set
				delete(missingSet, newHash)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Failed reading file:", path, err)
				break
			}
			if len(missingSet) == 0 {
				break // all blocks from this file restored
			}
		}
		f.Close()
	}

	// Check which blocks are still missing after repair
	stillMissing := []string{}
	for _, h := range missing {
		if err := r.Blocks.Get(h, io.Discard, true); err != nil {
			stillMissing = append(stillMissing, h)
		}
	}

	if len(stillMissing) > 0 {
		fmt.Printf("Blocks still missing after repair: %d\n", len(stillMissing))
		for _, h := range stillMissing {
			fmt.Println("  ", h)
		}
	} else {
		fmt.Println("All missing blocks restored successfully from workspace.")
	}

	return nil
}
