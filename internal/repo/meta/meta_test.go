package meta_test

import (
	"testing"
	"time"

	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/internal/repo/meta"
)

// writeCommit helper
func writeCommit(t *testing.T, m meta.MetaContextInterface, commit *meta.Commit) string {
	t.Helper()
	id, err := m.CreateCommit(commit)
	if err != nil {
		t.Fatalf("CreateCommit failed: %v", err)
	}
	return id
}

// prepare in-memory repo
func setupMemRepo(t *testing.T) *repo.Repository {
	t.Helper()
	memfs := fs.NewMemoryFS()

	// create the repo dirs manually, so CreateCommit won't panic
	memfs.MkdirAll("wt/.bvc/commits", 0o755)
	memfs.MkdirAll("wt/.bvc/branches", 0o755)
	memfs.MkdirAll("wt/.bvc/snapshots", 0o755)
	memfs.MkdirAll("wt/.bvc/blocks", 0o755)

	r, err := repo.NewRepositoryWithFS(&config.RepoConfig{RepoDir: "wt"}, memfs)
	if err != nil {
		t.Fatalf("NewRepositoryWithFS failed: %v", err)
	}
	return r
}

func TestInitAndOpenRepository(t *testing.T) {
	r := setupMemRepo(t)

	headData, err := r.FS.ReadFile(r.Config.HeadFile())
	if err != nil {
		t.Fatalf("failed to read HEAD: %v", err)
	}
	if string(headData) != "ref: branches/main" {
		t.Errorf("unexpected HEAD content: %s", string(headData))
	}
}

func TestBranchCreationAndListing(t *testing.T) {
	r := setupMemRepo(t)

	cur, err := r.Meta.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if cur.Name != config.DefaultBranch {
		t.Errorf("expected branch %s got %s", config.DefaultBranch, cur.Name)
	}

	newBranch, err := r.Meta.CreateBranch("feature")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}
	if newBranch.Name != "feature" {
		t.Errorf("expected branch feature got %s", newBranch.Name)
	}

	exists, err := r.Meta.BranchExists("feature")
	if err != nil {
		t.Fatalf("BranchExists failed: %v", err)
	}
	if !exists {
		t.Errorf("expected branch feature to exist")
	}

	branches, err := r.Meta.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("expected 2 branches got %d", len(branches))
	}
}

func TestCommitsLifecycle(t *testing.T) {
	r := setupMemRepo(t)

	branch := config.DefaultBranch
	commit := &meta.Commit{
		ID:        "abc123",
		Branch:    branch,
		Message:   "Initial commit",
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: "fileset1",
	}

	id := writeCommit(t, r.Meta, commit)
	if id != commit.ID {
		t.Errorf("expected ID %s got %s", commit.ID, id)
	}

	if err := r.Meta.SetLastCommitID(branch, commit.ID); err != nil {
		t.Fatalf("SetLastCommitID failed: %v", err)
	}

	lastID, err := r.Meta.GetLastCommitID(branch)
	if err != nil {
		t.Fatalf("GetLastCommitID failed: %v", err)
	}
	if lastID != commit.ID {
		t.Errorf("expected lastID %s got %s", commit.ID, lastID)
	}

	c, err := r.Meta.GetCommit(commit.ID)
	if err != nil {
		t.Fatalf("GetCommit failed: %v", err)
	}
	if c.Message != commit.Message {
		t.Errorf("expected message %q got %q", commit.Message, c.Message)
	}

	ids, err := r.Meta.AllCommitIDs(branch)
	if err != nil {
		t.Fatalf("AllCommitIDs failed: %v", err)
	}
	if len(ids) != 1 || ids[0] != commit.ID {
		t.Errorf("unexpected AllCommitIDs: %v", ids)
	}

	lastCommit, err := r.Meta.GetLastCommitForBranch(branch)
	if err != nil {
		t.Fatalf("GetLastCommitForBranch failed: %v", err)
	}
	if lastCommit.ID != commit.ID {
		t.Errorf("expected last commit ID %s got %s", commit.ID, lastCommit.ID)
	}
}

func TestHeadRefSetAndGet(t *testing.T) {
	r := setupMemRepo(t)

	ref, err := r.Meta.SetHeadRef("main")
	if err != nil {
		t.Fatalf("SetHeadRef failed: %v", err)
	}
	if ref.String() != "branches/main" {
		t.Errorf("unexpected HeadRef: %s", ref)
	}

	gotRef, err := r.Meta.GetHeadRef()
	if err != nil {
		t.Fatalf("GetHeadRef failed: %v", err)
	}
	if gotRef.String() != ref.String() {
		t.Errorf("HeadRef mismatch: expected %s got %s", ref, gotRef)
	}
}

func TestAllCommitIDsCycles(t *testing.T) {
	r := setupMemRepo(t)

	commitA := &meta.Commit{
		ID:        "A",
		Parents:   []string{"B"}, // cycle
		Branch:    config.DefaultBranch,
		Message:   "A",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	commitB := &meta.Commit{
		ID:        "B",
		Parents:   []string{"A"}, // cycle
		Branch:    config.DefaultBranch,
		Message:   "B",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	writeCommit(t, r.Meta, commitA)
	writeCommit(t, r.Meta, commitB)
	r.Meta.SetLastCommitID(config.DefaultBranch, "A")

	ids, err := r.Meta.AllCommitIDs(config.DefaultBranch)
	if err != nil {
		t.Fatalf("AllCommitIDs failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 commits due to cycle guard, got %d", len(ids))
	}
}
