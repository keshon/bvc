package analyze

import (
	"app/internal/command"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/repo"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Command struct{}

func (c *Command) Name() string      { return "analyze" }
func (c *Command) Short() string     { return "a" }
func (c *Command) Aliases() []string { return []string{"a"} }
func (c *Command) Usage() string     { return "analyze [--detail] [--export]" }
func (c *Command) Brief() string     { return "Analyze block reuse across the entire repository" }
func (c *Command) Help() string {
	return `Analyze block reuse across all branches and commits.

Usage:
  analyze --detail - print detailed shared block list
  analyze --export - save output to ${config.RepoDir}-analyze`
}

func (c *Command) Run(ctx *command.Context) error {
	full := false   // --detail
	export := false // --export

	for _, arg := range ctx.Args {
		switch arg {
		case "--full":
			full = true
		case "--export":
			export = true
		}
	}

	var exportBuf strings.Builder
	writeOut := func(s string) {
		fmt.Print(s)
		if export {
			exportBuf.WriteString(stripANSI(s))
		}
	}

	// open the repository context
	repo, err := repo.OpenAt(config.ResolveRepoRoot())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// list all branches
	branches, err := repo.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}
	if len(branches) == 0 {
		writeOut("No branches found.\n")
		return nil
	}

	blockFiles := map[string]map[string]struct{}{}    // hash -> path -> struct{}
	blockBranches := map[string]map[string]struct{}{} // hash -> branch -> struct{}
	blockCounts := map[string]int{}                   // hash -> count
	fileBlocks := map[string][]string{}               // path -> hashes

	// iterate over all branches
	for _, branch := range branches {
		lastCommit, err := repo.GetLastCommitForBranch(branch.Name)
		if err != nil || lastCommit == nil || lastCommit.ID == "" {
			continue
		}

		fileset, err := repo.GetCommitFileset(lastCommit.FilesetID)
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

	totalBlocks := len(blockCounts)
	sharedBlocks := 0
	for _, count := range blockCounts {
		if count > 1 {
			sharedBlocks++
		}
	}

	// file-level reuse overview
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

	// summary output
	writeOut("\033[96mSummary\033[0m\n\n")
	writeOut(prettyLine("\033[36mTotal branches\033[0m", fmt.Sprintf("%d", len(branches))) + "\n\n")
	writeOut(prettyLine("\033[36mTotal blocks\033[0m", fmt.Sprintf("%d", totalBlocks)) + "\n")
	writeOut(prettyLine("\033[36mUnique blocks\033[0m", fmt.Sprintf("%d", totalBlocks-sharedBlocks)) + "\n")
	writeOut(prettyLine("\033[36mShared blocks\033[0m", fmt.Sprintf("%d", sharedBlocks)) + "\n")
	writeOut(prettyLine("\033[36mOverall reuse ratio\033[0m", fmt.Sprintf("%.1f%%", float64(sharedBlocks)/float64(totalBlocks)*100)) + "\n\n")
	writeOut(prettyLine("\033[36mTotal files\033[0m", fmt.Sprintf("%d", totalFiles)) + "\n")
	writeOut(prettyLine("\033[36mFiles with shared blocks\033[0m", fmt.Sprintf("%d", filesWithShared)) + "\n")
	writeOut(prettyLine("\033[36mFile reuse ratio\033[0m", fmt.Sprintf("%.1f%%", fileSharedPercent)) + "\n")
	writeOut(prettyLine("\033[36mAvg. file reuse ratio\033[0m", fmt.Sprintf("%.1f%%", avgFileReuse)) + "\n\n")

	// if not full mode, stop here
	if !full {
		if export {
			saveExport(exportBuf.String())
		}
		return nil
	}

	// struct represents a row in output list
	type Row struct {
		Hash     string
		Files    []string
		Branches []string
		Count    int
	}

	// sort shared blocks
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

	sort.Slice(sharedList, func(i, j int) bool {
		return sharedList[i].Count > sharedList[j].Count
	})

	// write shared block list
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

func prettyLine(label string, value string) string {
	const width = 45
	dots := max(width-len(stripANSI(label)), 2)
	return fmt.Sprintf("%s\033[90m%s\033[0m %s", label, strings.Repeat(".", dots), value)
}

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func saveExport(content string) {
	filename := config.RepoDir + "-analyze"
	_ = os.WriteFile(filepath.Clean(filename), []byte(strings.TrimSpace(content)+"\n"), 0644)
	fmt.Printf("\n\033[90mExported analysis to %s\033[0m\n", filename)
}

func init() {
	command.RegisterCommand(
		command.ApplyMiddlewares(
			&Command{},
			middleware.WithDebugArgsPrint(),
		),
	)
}
