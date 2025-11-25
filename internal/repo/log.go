// repo/log.go
package repo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/keshon/bvc/internal/repo/meta"
)

type CommitFilter struct {
	AllBranches bool
	Branch      string
	Limit       int
	Since       *time.Time
	Until       *time.Time
}

type CommitWithRefs struct {
	Commit *meta.Commit
	Refs   []string
}

func (r *Repository) Log(filter CommitFilter) ([]*CommitWithRefs, error) {
	var branches []string

	// determine branches to scan
	if filter.Branch != "" {
		branches = []string{filter.Branch}
	} else if filter.AllBranches {
		all, err := r.Meta.ListBranches()
		if err != nil {
			return nil, fmt.Errorf("list branches: %w", err)
		}
		for _, b := range all {
			branches = append(branches, b.Name)
		}
	} else {
		cur, err := r.Meta.GetCurrentBranch()
		if err != nil {
			return nil, err
		}
		branches = []string{cur.Name}
	}

	seen := make(map[string]bool)
	var commits []*CommitWithRefs

	for _, branch := range branches {
		branchCommits, err := r.Meta.GetCommitsForBranch(branch)
		if err != nil {
			return nil, fmt.Errorf("get commits for branch %q: %w", branch, err)
		}

		for _, cmt := range branchCommits {
			if seen[cmt.ID] {
				continue
			}
			seen[cmt.ID] = true

			t, err := time.Parse(time.RFC3339, cmt.Timestamp)
			if err != nil {
				continue
			}
			if filter.Since != nil && t.Before(*filter.Since) {
				continue
			}
			if filter.Until != nil && t.After(*filter.Until) {
				continue
			}

			refs, _ := r.findRefsForCommit(cmt.ID)
			commits = append(commits, &CommitWithRefs{
				Commit: cmt,
				Refs:   refs,
			})
		}
	}

	// sort newest first
	sort.Slice(commits, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, commits[i].Commit.Timestamp)
		tj, _ := time.Parse(time.RFC3339, commits[j].Commit.Timestamp)
		return ti.After(tj)
	})

	if filter.Limit > 0 && filter.Limit < len(commits) {
		commits = commits[:filter.Limit]
	}

	return commits, nil
}

// helper: find branches where this commit is last commit
func (r *Repository) findRefsForCommit(commitID string) ([]string, error) {
	branches, err := r.Meta.ListBranches()
	if err != nil {
		return nil, err
	}

	var refs []string
	head, _ := r.Meta.GetCurrentBranch()

	for _, b := range branches {
		lastID, err := r.Meta.GetLastCommitID(b.Name)
		if err != nil {
			continue
		}
		if lastID == commitID {
			if head != nil && b.Name == head.Name {
				refs = append(refs, "HEAD -> "+b.Name)
			} else {
				refs = append(refs, b.Name)
			}
		}
	}

	// sort HEAD first, then alphabetical
	sort.Slice(refs, func(i, j int) bool {
		if strings.HasPrefix(refs[i], "HEAD ->") {
			return true
		}
		return refs[i] < refs[j]
	})

	return refs, nil
}
