package scan_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"

	"github.com/keshon/bvc/internal/repo/meta"
	"github.com/keshon/bvc/internal/repo/scan"
)

// Normalize test paths because MemoryFS does NOT support leading slashes
func p(path string) string {
	return filepath.Clean(path)
}

// Write JSON safely
func writeJSON(t *testing.T, memfs fs.FS, path string, v interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	if err := memfs.WriteFile(p(path), data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// Fake Meta
type fakeMeta struct {
	cfg     *config.RepoConfig
	branch  *meta.Branch
	commit  *meta.Commit
	commits map[string]*meta.Commit
}

func (f *fakeMeta) GetCurrentBranch() (*meta.Branch, error) { return f.branch, nil }
func (f *fakeMeta) GetBranch(name string) (meta.Branch, error) {
	if f.branch != nil && f.branch.Name == name {
		return *f.branch, nil
	}
	return meta.Branch{}, nil
}
func (f *fakeMeta) ListBranches() ([]meta.Branch, error)          { return nil, nil }
func (f *fakeMeta) CreateBranch(name string) (meta.Branch, error) { return meta.Branch{}, nil }
func (f *fakeMeta) BranchExists(name string) (bool, error)        { return false, nil }
func (f *fakeMeta) CreateBranchAt(name, commitID string, force bool) (meta.Branch, error) {
	return meta.Branch{}, nil
}
func (f *fakeMeta) DeleteBranch(name string) error                   { return nil }
func (f *fakeMeta) RenameBranch(old, new string, force bool) error   { return nil }
func (f *fakeMeta) CreateCommit(commit *meta.Commit) (string, error) { return "", nil }
func (f *fakeMeta) SetLastCommitID(branch, commitID string) error    { return nil }
func (f *fakeMeta) GetLastCommitID(branch string) (string, error)    { return "", nil }
func (f *fakeMeta) AllCommitIDs(branch string) ([]string, error)     { return nil, nil }
func (f *fakeMeta) GetConfig() *config.RepoConfig                    { return f.cfg }
func (f *fakeMeta) GetHeadRef() (meta.HeadRef, error)                { var hr meta.HeadRef; return hr, nil }
func (f *fakeMeta) SetHeadRef(branch string) (meta.HeadRef, error) {
	var hr meta.HeadRef
	return hr, nil
}
func (f *fakeMeta) GetCommit(commitID string) (*meta.Commit, error) {
	if f.commits == nil {
		return nil, nil
	}
	return f.commits[commitID], nil
}
func (f *fakeMeta) GetCommitsForBranch(branch string) ([]*meta.Commit, error) {
	if f.commit == nil {
		return nil, nil
	}
	return []*meta.Commit{f.commit}, nil
}
func (f *fakeMeta) GetLastCommitForBranch(branch string) (*meta.Commit, error) { return f.commit, nil }

// Test scanning a working tree with no commits and no index
func TestScanner_NoCommit_NoIndex(t *testing.T) {
	memfs := fs.NewMemoryFS()
	wt := "wt"
	memfs.MkdirAll(p(wt), 0o755)
	memfs.MkdirAll(p("wt/.bvc"), 0o755)
	memfs.WriteFile(p("wt/a.txt"), []byte("A"), 0o644)
	memfs.WriteFile(p("wt/b.txt"), []byte("B"), 0o644)

	m := &fakeMeta{cfg: &config.RepoConfig{RepoDir: "wt/.bvc"}, branch: &meta.Branch{Name: "main"}, commit: nil, commits: map[string]*meta.Commit{}}
	s := scan.NewScanner(wt, m, memfs)
	res, err := s.ScanAll()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(res.Tracked) != 0 {
		t.Fatalf("expected no tracked but got %#v", res.Tracked)
	}
	if len(res.Staged) != 0 {
		t.Fatalf("expected no staged")
	}
	if len(res.Untracked) != 2 {
		t.Fatalf("expected 2 untracked, got %#v", res.Untracked)
	}
}

// Test scanning staged files from index.json
func TestScanner_IndexStaging(t *testing.T) {
	memfs := fs.NewMemoryFS()
	wt := "wt"
	memfs.MkdirAll(p(wt), 0o755)
	memfs.MkdirAll(p("wt/.bvc"), 0o755)
	memfs.WriteFile(p("wt/a.txt"), []byte("A"), 0o644)
	memfs.WriteFile(p("wt/b.txt"), []byte("B"), 0o644)
	writeJSON(t, memfs, "wt/.bvc/index.json", []map[string]string{{"path": "a.txt"}})

	m := &fakeMeta{
		cfg:    &config.RepoConfig{RepoDir: "wt/.bvc"},
		branch: &meta.Branch{Name: "main"},
		commit: nil, commits: map[string]*meta.Commit{},
	}
	s := scan.NewScanner(wt, m, memfs)
	res, err := s.ScanAll()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(res.Staged) != 1 || res.Staged[0] != "a.txt" {
		t.Fatalf("expected staged=[a.txt], got %#v", res.Staged)
	}
	if len(res.Untracked) != 1 || res.Untracked[0] != "b.txt" {
		t.Fatalf("expected untracked=[b.txt], got %#v", res.Untracked)
	}
}

// Test scanning files tracked by last commit
func TestScanner_TrackedFromCommit(t *testing.T) {
	memfs := fs.NewMemoryFS()
	memfs.MkdirAll(p("wt"), 0o755)
	memfs.MkdirAll(p("wt/.bvc"), 0o755)
	memfs.MkdirAll(p("wt/.bvc/snapshots"), 0o755)
	memfs.WriteFile(p("wt/a.txt"), []byte("A"), 0o644)
	writeJSON(t, memfs, "wt/.bvc/snapshots/s1.json", map[string]interface{}{"files": []map[string]interface{}{{"path": "a.txt"}}})

	commit := &meta.Commit{ID: "s1", FilesetID: "s1", Parents: []string{}}
	m := &fakeMeta{
		cfg:     &config.RepoConfig{RepoDir: "wt/.bvc"},
		branch:  &meta.Branch{Name: "main"},
		commit:  commit,
		commits: map[string]*meta.Commit{"s1": commit},
	}
	s := scan.NewScanner("wt", m, memfs)
	res, err := s.ScanAll()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(res.Tracked) != 1 || res.Tracked[0] != "a.txt" {
		t.Fatalf("expected tracked=[a.txt], got %#v", res.Tracked)
	}
}

// Test scanning ignored files
func TestScanner_Ignored(t *testing.T) {
	memfs := fs.NewMemoryFS()
	memfs.MkdirAll(p("wt"), 0o755)
	memfs.MkdirAll(p("wt/.bvc"), 0o755)
	memfs.WriteFile(p("wt/a.log"), []byte("X"), 0o644)
	memfs.WriteFile(p("wt/b.txt"), []byte("Y"), 0o644)
	memfs.WriteFile(p("wt/.bvc-ignore"), []byte("*.log\n"), 0o644)

	m := &fakeMeta{
		cfg:    &config.RepoConfig{RepoDir: "wt/.bvc"},
		branch: &meta.Branch{Name: "main"},
		commit: nil, commits: map[string]*meta.Commit{},
	}
	s := scan.NewScanner("wt", m, memfs)
	res, err := s.ScanAll()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(res.Ignored) != 1 || res.Ignored[0] != "a.log" {
		t.Fatalf("expected ignored=[a.log], got %#v", res.Ignored)
	}
	if len(res.Untracked) != 1 || res.Untracked[0] != "b.txt" {
		t.Fatalf("expected untracked=[b.txt], got %#v", res.Untracked)
	}
}

// Test scanning nested directories with mixed tracked, untracked, and ignored files
func TestScanner_Nested(t *testing.T) {
	memfs := fs.NewMemoryFS()
	memfs.MkdirAll(p("wt/dir/sub"), 0o755)
	memfs.MkdirAll(p("wt/.bvc"), 0o755)
	memfs.MkdirAll(p("wt/.bvc/snapshots"), 0o755)
	memfs.WriteFile(p("wt/dir/x.txt"), []byte("X"), 0o644)
	memfs.WriteFile(p("wt/dir/sub/y.tmp"), []byte("Y"), 0o644)
	memfs.WriteFile(p("wt/z.txt"), []byte("Z"), 0o644)
	memfs.WriteFile(p("wt/.bvc-ignore"), []byte("*.tmp\n"), 0o644)
	writeJSON(t, memfs, "wt/.bvc/snapshots/s1.json", map[string]interface{}{"files": []map[string]interface{}{{"path": "z.txt"}}})

	commit := &meta.Commit{ID: "s1", FilesetID: "s1", Parents: []string{}}
	m := &fakeMeta{
		cfg:     &config.RepoConfig{RepoDir: "wt/.bvc"},
		branch:  &meta.Branch{Name: "main"},
		commit:  commit,
		commits: map[string]*meta.Commit{"s1": commit},
	}
	s := scan.NewScanner("wt", m, memfs)
	res, err := s.ScanAll()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	expect := func(label string, got []string, want []string) {
		if len(got) != len(want) {
			t.Fatalf("%s expected %#v, got %#v", label, want, got)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%s mismatch: expected %#v, got %#v", label, want, got)
			}
		}
	}

	expect("tracked", res.Tracked, []string{"z.txt"})
	expect("ignored", res.Ignored, []string{"dir/sub/y.tmp"})
	expect("untracked", res.Untracked, []string{"dir/x.txt"})
}
