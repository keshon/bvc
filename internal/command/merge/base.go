package merge

import (
	"app/internal/config"
	"app/internal/repo"
	"app/internal/storage/file"
	"app/internal/storage/snapshot"

	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/zeebo/xxh3"
)

// findCommonAncestor walks commit history to find merge base.
// Returns commit ID or error if no common ancestor found.
func findCommonAncestor(aCommitID, bCommitID string) (string, error) {
	if aCommitID == "" || bCommitID == "" {
		return "", nil
	}

	// Open the repository context
	r, err := repo.OpenAt(config.RepoDir)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}
	seen := map[string]bool{}
	// walk a's ancestors (including a)
	var stack []string
	stack = append(stack, aCommitID)
	for len(stack) > 0 {
		commitID := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if commitID == "" || seen[commitID] {
			continue
		}
		seen[commitID] = true

		var c *repo.Commit
		c, err := r.GetCommit(commitID)
		if err != nil {
			continue
		}

		for _, p := range c.Parents {
			if p != "" {
				stack = append(stack, p)
			}
		}
	}

	// walk b's ancestors breadth-first and find first that exists in seen
	queue := []string{bCommitID}
	visited := map[string]bool{}
	for len(queue) > 0 {
		commitID := queue[0]
		queue = queue[1:]
		if commitID == "" || visited[commitID] {
			continue
		}
		visited[commitID] = true
		if seen[commitID] {
			return commitID, nil
		}

		var c *repo.Commit
		c, err := r.GetCommit(commitID)
		if err != nil {
			continue
		}

		for _, p := range c.Parents {
			if p != "" {
				queue = append(queue, p)
			}
		}
	}

	return "", nil // no common ancestor

}

// mergeFilesets performs three-way merge of filesets.
// Returns merged fileset and list of conflicting paths.
func mergeFilesets(base, ours, theirs *snapshot.Fileset) (snapshot.Fileset, []string) {
	// returns merged fileset and list of conflict paths
	conflicts := []string{}
	mergedMap := map[string]file.Entry{}

	// Create maps for quick lookup
	baseMap := map[string]file.Entry{}
	for _, f := range base.Files {
		baseMap[filepath.Clean(f.Path)] = f
	}
	oursMap := map[string]file.Entry{}
	for _, f := range ours.Files {
		oursMap[filepath.Clean(f.Path)] = f
	}
	theirsMap := map[string]file.Entry{}
	for _, f := range theirs.Files {
		theirsMap[filepath.Clean(f.Path)] = f
	}

	// union of all paths
	allPaths := map[string]bool{}
	for p := range baseMap {
		allPaths[p] = true
	}
	for p := range oursMap {
		allPaths[p] = true
	}
	for p := range theirsMap {
		allPaths[p] = true
	}

	for path := range allPaths {
		var b *file.Entry
		var o *file.Entry
		var t *file.Entry

		if v, ok := baseMap[path]; ok {
			tmp := v
			b = &tmp
		}
		if v, ok := oursMap[path]; ok {
			o = &v
		}
		if v, ok := theirsMap[path]; ok {
			t = &v
		}

		// Cases
		switch {
		// identical theirs and ours -> take either (or nothing if both nil)
		case o.Equal(t):
			if o != nil {
				mergedMap[path] = *o
			}
			// if both nil -> deleted in both -> skip

		// unchanged in ours (base == ours) -> take theirs (could be add, modify, or delete)
		case b.Equal(o):
			if t != nil {
				mergedMap[path] = *t
			} else {
				// theirs deleted -> delete in merged
				// i.e. do nothing (omit from mergedMap)
			}

		// unchanged in theirs (base == theirs) -> take ours
		case b.Equal(t):
			if o != nil {
				mergedMap[path] = *o
			} else {
				// ours deleted -> deleted in merged
			}

		// conflict: both changed differently since base (or base nil and both changed differently)
		default:
			// Conflict resolution policy: keep ours, write theirs to .MERGE_THEIRS
			if o != nil {
				mergedMap[path] = *o
			}
			if t != nil {
				// create duplicate path for their version
				conflictPath := path + ".MERGE_THEIRS"
				theirsCopy := *t
				theirsCopy.Path = conflictPath
				mergedMap[conflictPath] = theirsCopy
				conflicts = append(conflicts, path)
			}
		}
	}

	// build Fileset struct
	mergedFiles := make([]file.Entry, 0, len(mergedMap))
	for _, f := range mergedMap {
		mergedFiles = append(mergedFiles, f)
	}
	// deterministic order
	sort.SliceStable(mergedFiles, func(i, j int) bool {
		return filepath.Clean(mergedFiles[i].Path) < filepath.Clean(mergedFiles[j].Path)
	})

	filesetID := snapshot.HashFileset(mergedFiles)
	return snapshot.Fileset{ID: filesetID, Files: mergedFiles}, conflicts
}

// merge executes a full merge operation between branches.
func merge(currentBranch, targetBranch string) error {
	// basic checks
	if currentBranch == targetBranch {
		return fmt.Errorf("cannot merge branch into itself")
	}

	// Open the repository context
	r, err := repo.OpenAt(config.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// get commits
	currentCommitID, _ := r.GetLastCommitID(currentBranch)
	targetCommitID, err := r.GetLastCommitID(targetBranch)
	if err != nil {
		return err
	}
	if targetCommitID == "" {
		return fmt.Errorf("branch %s has no commits", targetBranch)
	}

	// find base
	baseID, err := findCommonAncestor(currentCommitID, targetCommitID)
	if err != nil {
		return err
	}
	if baseID == "" {
		return fmt.Errorf("no common ancestor found between '%s' and '%s'", currentBranch, targetBranch)
	}

	// load filesets
	baseFS, err := r.GetCommitFileset(baseID)
	if err != nil {
		return fmt.Errorf("failed to load base fileset: %v", err)
	}
	oursFS, err := r.GetCommitFileset(currentCommitID)
	if err != nil {
		return fmt.Errorf("failed to load our fileset: %v", err)
	}
	theirsFS, err := r.GetCommitFileset(targetCommitID)
	if err != nil {
		return fmt.Errorf("failed to load their fileset: %v", err)
	}

	// perform three-way merge
	mergedFS, conflicts := mergeFilesets(baseFS, oursFS, theirsFS)

	// save merged fileset
	r.Storage.Snapshots.Save(mergedFS)

	// create merge commit with two parents
	hash128 := xxh3.Hash128([]byte(
		mergedFS.ID + currentCommitID + targetCommitID + time.Now().String(),
	)).Bytes()
	commitID := fmt.Sprintf("%x", hash128[:8])

	mergeCommit := repo.Commit{
		ID:        commitID,
		Parents:   []string{currentCommitID, targetCommitID},
		Branch:    currentBranch,
		Message:   fmt.Sprintf("Merge branch '%s' into '%s'", targetBranch, currentBranch),
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: mergedFS.ID,
	}

	// create merge commit
	_, err = r.CreateCommit(&mergeCommit)
	if err != nil {
		return fmt.Errorf("failed to create merge commit: %v", err)
	}

	// update current branch to point to new merge commit
	if err := r.SetLastCommitID(currentBranch, commitID); err != nil {
		return fmt.Errorf("failed to update branch: %v", err)
	}

	// apply merged fileset to working directory
	if err := r.Storage.Files.Restore(mergedFS.Files, fmt.Sprintf("merge of %s", targetBranch)); err != nil {
		return fmt.Errorf("failed to apply merged fileset: %v", err)
	}

	// report conflicts
	if len(conflicts) > 0 {
		fmt.Println("\nMerge completed with conflicts:")
		for _, path := range conflicts {
			fmt.Printf("CONFLICT: %s (theirs saved as %s.MERGE_THEIRS)\n", path, path)
		}
		fmt.Println("\nResolve conflicts manually and commit the result.")
	} else {
		fmt.Printf("\nMerge completed successfully: '%s' merged into '%s'\n", targetBranch, currentBranch)
	}

	return nil
}
