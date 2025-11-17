package block

import (
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo"
)

type ReuseCommand struct {
	full   bool
	export bool
}

func (c *ReuseCommand) Name() string                   { return "reuse" }
func (c *ReuseCommand) Brief() string                  { return "Analyze block reuse across the repo" }
func (c *ReuseCommand) Usage() string                  { return "block reuse [--full] [--export]" }
func (c *ReuseCommand) Help() string                   { return "Analyze block reuse across branches" }
func (c *ReuseCommand) Aliases() []string              { return []string{"a"} }
func (c *ReuseCommand) Subcommands() []command.Command { return nil }
func (c *ReuseCommand) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.full, "full", false, "Print detailed shared block list")
	fs.BoolVar(&c.export, "export", false, "Save output to file")
}

func (c *ReuseCommand) Run(ctx *command.Context) error {
	full := c.full
	export := c.export

	var exportBuf strings.Builder
	writeOut := func(s string) {
		fmt.Print(s)
		if export {
			exportBuf.WriteString(stripANSI(s))
		}
	}

	r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	branches, err := r.Meta.ListBranches()
	if err != nil || len(branches) == 0 {
		writeOut("No branches found.\n")
		return nil
	}

	blockFiles := map[string]map[string]struct{}{}
	blockBranches := map[string]map[string]struct{}{}
	blockCounts := map[string]int{}
	fileBlocks := map[string][]string{}

	for _, branch := range branches {
		lastCommit, err := r.Meta.GetLastCommitForBranch(branch.Name)
		if err != nil || lastCommit == nil || lastCommit.ID == "" {
			continue
		}

		fileset, err := r.GetCommittedFileset(lastCommit.ID)

		if err != nil {
			continue
		}

		for _, file := range fileset.Files {
			for _, block := range file.Blocks {
				blockCounts[block.Hash]++
				if _, ok := blockFiles[block.Hash]; !ok {
					blockFiles[block.Hash] = map[string]struct{}{}
				}
				blockFiles[block.Hash][file.Path] = struct{}{}

				if _, ok := blockBranches[block.Hash]; !ok {
					blockBranches[block.Hash] = map[string]struct{}{}
				}
				blockBranches[block.Hash][branch.Name] = struct{}{}

				fileBlocks[file.Path] = append(fileBlocks[file.Path], block.Hash)
			}
		}
	}

	// Summary output
	totalBlocks := len(blockCounts)
	sharedBlocks := 0
	for _, count := range blockCounts {
		if count > 1 {
			sharedBlocks++
		}
	}

	totalFiles := len(fileBlocks)
	filesWithShared := 0
	totalFileReuseRatio := 0.0
	for _, blockHashes := range fileBlocks {
		shared := 0
		for _, h := range blockHashes {
			if blockCounts[h] > 1 {
				shared++
			}
		}
		if shared > 0 {
			filesWithShared++
		}
		if len(blockHashes) > 0 {
			totalFileReuseRatio += float64(shared) / float64(len(blockHashes))
		}
	}

	fileSharedPercent := 0.0
	if totalFiles > 0 {
		fileSharedPercent = float64(filesWithShared) / float64(totalFiles) * 100
	}

	avgFileReuse := 0.0
	if totalFiles > 0 {
		avgFileReuse = totalFileReuseRatio / float64(totalFiles) * 100
	}

	writeOut("\033[96mSummary\033[0m\n\n")
	writeOut(prettyLine("\033[36mTotal branches\033[0m", fmt.Sprintf("%d", len(branches))) + "\n")
	writeOut(prettyLine("\033[36mTotal blocks\033[0m", fmt.Sprintf("%d", totalBlocks)) + "\n")
	writeOut(prettyLine("\033[36mUnique blocks\033[0m", fmt.Sprintf("%d", totalBlocks-sharedBlocks)) + "\n")
	writeOut(prettyLine("\033[36mShared blocks\033[0m", fmt.Sprintf("%d", sharedBlocks)) + "\n")
	writeOut(prettyLine("\033[36mOverall reuse ratio\033[0m", fmt.Sprintf("%.1f%%", float64(sharedBlocks)/float64(totalBlocks)*100)) + "\n")
	writeOut(prettyLine("\033[36mFiles with shared blocks\033[0m", fmt.Sprintf("%d", filesWithShared)) + "\n")
	writeOut(prettyLine("\033[36mFile reuse ratio\033[0m", fmt.Sprintf("%.1f%%", fileSharedPercent)) + "\n")
	writeOut(prettyLine("\033[36mAvg. file reuse ratio\033[0m", fmt.Sprintf("%.1f%%", avgFileReuse)) + "\n\n")

	if !full {
		if export {
			saveExport(exportBuf.String())
		}
		return nil
	}

	// Full detailed output
	type Row struct {
		Hash     string
		Files    []string
		Branches []string
		Count    int
	}
	var sharedList []Row
	for hash, count := range blockCounts {
		if count <= 1 {
			continue
		}
		files := make([]string, 0, len(blockFiles[hash]))
		for f := range blockFiles[hash] {
			files = append(files, f)
		}
		sort.Strings(files)

		branches := make([]string, 0, len(blockBranches[hash]))
		for b := range blockBranches[hash] {
			branches = append(branches, b)
		}
		sort.Strings(branches)

		sharedList = append(sharedList, Row{
			Hash:     hash,
			Files:    files,
			Branches: branches,
			Count:    count,
		})
	}

	sort.Slice(sharedList, func(i, j int) bool { return sharedList[i].Count > sharedList[j].Count })

	writeOut("\n\033[96mShared Blocks (most reused first):\033[0m\n")
	if len(sharedList) == 0 {
		writeOut("  None\n")
	}
	for i, sb := range sharedList {
		writeOut(fmt.Sprintf("\n\033[36m[%d] %s\033[0m\n", i+1, sb.Hash))
		writeOut(fmt.Sprintf("  Occurrences: %d\n", sb.Count))
		writeOut(fmt.Sprintf("  Branches:    %s\n", strings.Join(sb.Branches, ", ")))
		writeOut("  Files:\n")
		for _, f := range sb.Files {
			writeOut(fmt.Sprintf("    - %s\n", f))
		}
	}

	if export {
		saveExport(exportBuf.String())
	}

	return nil
}

// helpers
func prettyLine(label, value string) string {
	const width = 45
	dots := max(width-len(stripANSI(label)), 2)
	return fmt.Sprintf("%s\033[90m%s\033[0m %s", label, strings.Repeat(".", dots), value)
}

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func saveExport(content string) {
	fs := fs.NewOSFS()
	filename := config.RepoDir + "-reuse"
	_ = fs.WriteFile(filepath.Clean(filename), []byte(strings.TrimSpace(content)+"\n"), 0644)
	fmt.Printf("\n\033[90mExported analysis to %s\033[0m\n", filename)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
