package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"app/internal/cli"
	"app/internal/config"
	"app/internal/core"
	"app/internal/progress"
	"app/internal/repo"
	"app/internal/storage/block"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"
	"app/internal/util"
)

// AnalyzeCommand inspects files and reports block reuse potential.
type AnalyzeCommand struct{}

func (c *AnalyzeCommand) Name() string { return "analyze" }
func (c *AnalyzeCommand) Usage() string {
	return "analyze <path|.> [--compare <commit>] [--by-type] [--json] [--dry-run] [--top N]"
}
func (c *AnalyzeCommand) Description() string {
	return "Analyze block reuse potential for files (estimate new data to store)"
}
func (c *AnalyzeCommand) DetailedDescription() string {
	return `Analyze files and predict how many blocks will be reused from existing repository objects.
Options (order doesn't matter):
  --compare <commit>   : compare against specific commit instead of all repo blocks
  --by-type            : group results by file extension and show best/worst types
  --json               : output machine-readable JSON
  --dry-run            : don't access/modify storage (analysis only)
  --top N              : show top N most/least reusable files (default 10)`
}
func (c *AnalyzeCommand) Aliases() []string { return nil }
func (c *AnalyzeCommand) Short() string     { return "A" }

// Run parses args in any order and executes analysis.
func (c *AnalyzeCommand) Run(ctx *cli.Context) error {
	// parse args/flags in any order
	var target string = "."
	var compareCommit string
	var byType bool
	var asJSON bool
	var dryRun bool
	topN := 10

	// parse ctx.Args for flags (supports both --flag=value and positional)
	for i := 0; i < len(ctx.Args); i++ {
		a := ctx.Args[i]
		switch {
		case a == "--compare" && i+1 < len(ctx.Args):
			compareCommit = ctx.Args[i+1]
			i++
		case strings.HasPrefix(a, "--compare="):
			compareCommit = strings.TrimPrefix(a, "--compare=")
		case a == "--by-type":
			byType = true
		case a == "--json":
			asJSON = true
		case a == "--dry-run":
			dryRun = true
		case a == "--top" && i+1 < len(ctx.Args):
			var n int
			_, err := fmt.Sscanf(ctx.Args[i+1], "%d", &n)
			if err == nil && n > 0 {
				topN = n
			}
			i++
		case strings.HasPrefix(a, "--top="):
			var n int
			_, err := fmt.Sscanf(strings.TrimPrefix(a, "--top="), "%d", &n)
			if err == nil && n > 0 {
				topN = n
			}
		default:
			// treat first non-flag as target path
			if !strings.HasPrefix(a, "-") && target == "." {
				target = a
			}
		}
	}

	// also respect flags parsed by CLI package (ctx.Flags)
	if val, ok := ctx.Flags["compare"]; ok {
		compareCommit = val
	}
	if _, ok := ctx.Flags["by-type"]; ok {
		byType = true
	}
	if _, ok := ctx.Flags["json"]; ok {
		asJSON = true
	}
	if _, ok := ctx.Flags["dry-run"]; ok {
		dryRun = true
	}
	if val, ok := ctx.Flags["top"]; ok {
		var n int
		_, err := fmt.Sscanf(val, "%d", &n)
		if err == nil && n > 0 {
			topN = n
		}
	}

	return runAnalyze(target, compareCommit, byType, asJSON, dryRun, topN)
}

// per-file stats
type fileStat struct {
	Path        string
	TotalBlocks int
	KnownBlocks int
	NewBlocks   int
	TotalBytes  int64
	KnownBytes  int64
	NewBytes    int64
	Ext         string
}

// output shape for JSON
type analyzeResult struct {
	Timestamp     string              `json:"timestamp"`
	Target        string              `json:"target"`
	CompareCommit string              `json:"compare_commit,omitempty"`
	TotalFiles    int                 `json:"total_files"`
	TotalBlocks   int                 `json:"total_blocks"`
	KnownBlocks   int                 `json:"known_blocks"`
	NewBlocks     int                 `json:"new_blocks"`
	TotalBytes    int64               `json:"total_bytes"`
	KnownBytes    int64               `json:"known_bytes"`
	NewBytes      int64               `json:"new_bytes"`
	ByType        map[string]typeStat `json:"by_type,omitempty"`
	TopReusable   []fileStat          `json:"top_reusable"`
	TopNew        []fileStat          `json:"top_new"`
}

type typeStat struct {
	Files      int    `json:"files"`
	TotalBytes int64  `json:"total_bytes"`
	ReusePct   string `json:"reuse_pct"`
}

// runAnalyze does the heavy lifting
func runAnalyze(target, compareCommit string, byType, asJSON, dryRun bool, topN int) error {
	start := time.Now()

	// collect repo blocks (global set)
	var repoBlocks map[string]*repo.BlockInfo
	var err error
	if compareCommit == "" {
		// Collect all blocks referenced by HEADs of all branches
		repoBlocks, err = repo.CollectAllBlocks()
		if err != nil {
			return fmt.Errorf("collect repo blocks: %w", err)
		}
	} else {
		// Load a specific commit's fileset blocks (only that commit)
		repoBlocks = make(map[string]*repo.BlockInfo)
		commit, err := core.GetCommit(compareCommit)
		if err != nil {
			return fmt.Errorf("unknown commit to compare: %s", compareCommit)
		}
		fsPath := filepath.Join(config.FilesetsDir, commit.FilesetID+".json")
		var fs snapshot.Fileset
		if err := util.ReadJSON(fsPath, &fs); err != nil {
			return fmt.Errorf("read fileset for commit %s: %w", compareCommit, err)
		}
		for _, f := range fs.Files {
			for _, b := range f.Blocks {
				repoBlocks[b.Hash] = &repo.BlockInfo{Size: b.Size}
			}
		}
	}

	// also create a quick lookup of actual object files present in objects dir (cache)
	objectsPresent := make(map[string]struct{})
	_ = filepath.WalkDir(config.ObjectsDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		n := d.Name()
		if strings.HasSuffix(n, ".bin") {
			hash := strings.TrimSuffix(n, ".bin")
			objectsPresent[hash] = struct{}{}
		}
		return nil
	})

	// collect files to analyze
	var files []string
	if target == "." {
		files, err = file.ListAll()
		if err != nil {
			return err
		}
	} else {
		// if path is a file -> single file; if dir -> walk
		info, err := os.Stat(target)
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = filepath.WalkDir(target, func(p string, d os.DirEntry, err error) error {
				if err != nil || d == nil {
					return nil
				}
				if d.IsDir() {
					// skip repo dir
					if filepath.Clean(p) == filepath.Clean(config.RepoDir) {
						return filepath.SkipDir
					}
					return nil
				}
				files = append(files, p)
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			files = []string{target}
		}
	}

	if len(files) == 0 {
		fmt.Println("No files found to analyze.")
		return nil
	}

	// progress bar
	bar := progress.NewProgress(len(files), "Analyzing files")
	defer bar.Finish()

	stats := make([]fileStat, 0, len(files))

	totalBlocks := 0
	knownBlocks := 0
	var totalBytes int64
	var knownBytes int64

	for _, f := range files {
		// skip index, repo internals
		clean := filepath.Clean(f)
		if strings.HasPrefix(clean, config.RepoDir+string(os.PathSeparator)) {
			bar.Increment()
			continue
		}

		blocks, err := block.SplitFileIntoBlocks(f)
		if err != nil {
			// skip unreadable files
			bar.Increment()
			continue
		}
		fs, _ := os.Stat(f)
		var size int64
		if fs != nil {
			size = fs.Size()
		}
		fsCount := len(blocks)
		fileKnown := 0
		var fileKnownBytes int64
		for _, b := range blocks {
			// known if present in repoBlocks OR objectsPresent
			if _, ok := repoBlocks[b.Hash]; ok {
				fileKnown++
				fileKnownBytes += b.Size
			} else if _, ok := objectsPresent[b.Hash]; ok {
				fileKnown++
				fileKnownBytes += b.Size
			}
		}
		fileNew := fsCount - fileKnown

		st := fileStat{
			Path:        f,
			TotalBlocks: fsCount,
			KnownBlocks: fileKnown,
			NewBlocks:   fileNew,
			TotalBytes:  size,
			KnownBytes:  fileKnownBytes,
			NewBytes:    size - fileKnownBytes,
			Ext:         strings.ToLower(filepath.Ext(f)),
		}
		stats = append(stats, st)

		totalBlocks += fsCount
		knownBlocks += fileKnown
		totalBytes += size
		knownBytes += fileKnownBytes

		bar.Increment()
	}

	// prepare aggregated result
	newBlocks := totalBlocks - knownBlocks
	newBytes := totalBytes - knownBytes
	reusePct := 0.0
	if totalBlocks > 0 {
		reusePct = float64(knownBlocks) / float64(totalBlocks) * 100.0
	}

	// sort for top reusable / most new
	sort.Slice(stats, func(i, j int) bool {
		// prefer higher reuse ratio
		ri := reuseRatio(stats[i])
		rj := reuseRatio(stats[j])
		if ri == rj {
			// fallback: larger files first
			return stats[i].TotalBytes > stats[j].TotalBytes
		}
		return ri > rj
	})

	topReusable := topNSlice(stats, topN, true)
	// for most new, sort ascending reuse
	sort.Slice(stats, func(i, j int) bool {
		return reuseRatio(stats[i]) < reuseRatio(stats[j])
	})
	topNew := topNSlice(stats, topN, true)

	// by-type aggregation
	var byTypeMap map[string]typeStat
	if byType {
		byTypeMap = map[string]typeStat{}
		for _, s := range stats {
			t := s.Ext
			ts := byTypeMap[t]
			ts.Files++
			ts.TotalBytes += s.TotalBytes
			byTypeMap[t] = ts
		}
		// compute reuse percent per type (approx) by re-scanning entries by ext
		// build map ext -> known bytes / total bytes
		typeKnown := map[string]int64{}
		typeTotal := map[string]int64{}
		for _, s := range stats {
			typeKnown[s.Ext] += s.KnownBytes
			typeTotal[s.Ext] += s.TotalBytes
		}
		for ext, ts := range byTypeMap {
			k := typeKnown[ext]
			tot := typeTotal[ext]
			pct := "N/A"
			if tot > 0 {
				pct = fmt.Sprintf("%.1f%%", float64(k)/float64(tot)*100.0)
			}
			ts.ReusePct = pct
			byTypeMap[ext] = ts
		}
	}

	// output
	res := analyzeResult{
		Timestamp:     start.Format(time.RFC3339),
		Target:        target,
		CompareCommit: compareCommit,
		TotalFiles:    len(stats),
		TotalBlocks:   totalBlocks,
		KnownBlocks:   knownBlocks,
		NewBlocks:     newBlocks,
		TotalBytes:    totalBytes,
		KnownBytes:    knownBytes,
		NewBytes:      newBytes,
		ByType:        byTypeMap,
		TopReusable:   topReusable,
		TopNew:        topNew,
	}

	if asJSON {
		out, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(out))
		return nil
	}

	// Human readable print
	fmt.Printf("Analyzed: %s\n", target)
	fmt.Printf("Compared against: %s\n", compareLabel(compareCommit))
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("Files analyzed:     %d\n", res.TotalFiles)
	fmt.Printf("Total blocks:       %d\n", res.TotalBlocks)
	fmt.Printf("Known blocks:       %d\n", res.KnownBlocks)
	fmt.Printf("New blocks:         %d\n", res.NewBlocks)
	fmt.Printf("Reusability:        %.1f%%\n", reusePct)
	fmt.Printf("Total bytes:        %s\n", humanBytes(res.TotalBytes))
	fmt.Printf("Estimated new data: %s\n", humanBytes(res.NewBytes))
	fmt.Println("────────────────────────────────────────────────────────────")

	if byType {
		fmt.Println("By extension:")
		// sort extensions by total bytes desc
		exts := make([]string, 0, len(byTypeMap))
		for e := range byTypeMap {
			exts = append(exts, e)
		}
		sort.Slice(exts, func(i, j int) bool {
			return byTypeMap[exts[i]].TotalBytes > byTypeMap[exts[j]].TotalBytes
		})
		for _, e := range exts {
			ts := byTypeMap[e]
			if e == "" {
				e = "(no ext)"
			}
			fmt.Printf("  %-10s  files:%4d  bytes:%8s  reuse:%6s\n", e, ts.Files, humanBytes(ts.TotalBytes), ts.ReusePct)
		}
		fmt.Println("────────────────────────────────────────────────────────────")
	}

	fmt.Printf("Top %d most reusable files (high reuse first):\n", len(topReusable))
	for _, s := range topReusable {
		fmt.Printf("  %6.1f%%  %8s  %s\n", reuseRatio(s)*100.0, humanBytes(s.TotalBytes), s.Path)
	}
	fmt.Println()
	fmt.Printf("Top %d least reusable files (low reuse first):\n", len(topNew))
	for _, s := range topNew {
		fmt.Printf("  %6.1f%%  %8s  %s\n", reuseRatio(s)*100.0, humanBytes(s.TotalBytes), s.Path)
	}
	fmt.Println()

	if dryRun {
		fmt.Println("Note: dry-run set — no storage modifications were performed.")
	}

	return nil
}

func compareLabel(commit string) string {
	if commit == "" {
		return "all repo blocks"
	}
	return commit
}

func reuseRatio(s fileStat) float64 {
	if s.TotalBlocks == 0 {
		return 0.0
	}
	return float64(s.KnownBlocks) / float64(s.TotalBlocks)
}

func topNSlice(stats []fileStat, n int, shallow bool) []fileStat {
	if n <= 0 {
		return nil
	}
	if len(stats) < n {
		n = len(stats)
	}
	out := make([]fileStat, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, stats[i])
	}
	return out
}

// humanBytes prints friendly sizes
func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(b) / float64(div)
	prefix := "KMGTPE"[exp : exp+1]
	return fmt.Sprintf("%.1f %sB", value, prefix)
}

func init() {
	cli.RegisterCommand(&AnalyzeCommand{})
}
