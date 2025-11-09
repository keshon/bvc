package block

import (
	"app/internal/config"
	"app/internal/fsio"
	"app/internal/repo/store/block"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/zeebo/xxh3"
)

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
			fmt.Printf("\033[31m%s\033[0m  files: %v  branches: %v\n", bc.Hash, files, bc.Branches)
		}
	}
	return failed
}

func verifyBlockHash(path, expected string) (bool, error) {
	data, err := fsio.ReadFile(path)
	if err != nil {
		return false, err
	}
	sum := fmt.Sprintf("%x", xxh3.Hash128(data).Bytes())
	return sum == expected, nil
}
