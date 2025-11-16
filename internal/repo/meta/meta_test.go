package meta_test

import (
	"os"
	"testing"
	"time"

	"app/internal/config"

	"app/internal/repo"
	"app/internal/repo/meta"
)

// helpers
var (
	ReadFileFunc  = os.ReadFile
	WriteFileFunc = os.WriteFile
)

func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bvc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir
}

// Init
func TestInitAndOpenRepository(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	if r.Meta.Config.RepoRoot != tmp {
		t.Errorf("expected Root=%q got %q", tmp, r.Meta.Config.RepoRoot)
	}

	// Check HEAD file
	headData, err := os.ReadFile(r.Config.HeadFile())
	if err != nil {
		t.Fatalf("failed to read HEAD: %v", err)
	}
	if string(headData) != "ref: branches/main" {
		t.Errorf("unexpected HEAD content: %s", string(headData))
	}

}

// Branches
func TestBranchCreationAndListing(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	// Current branch
	cur, err := r.Meta.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if cur.Name != config.DefaultBranch {
		t.Errorf("expected branch %s got %s", config.DefaultBranch, cur.Name)
	}

	// Create new branch
	newBranch, err := r.Meta.CreateBranch("feature")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}
	if newBranch.Name != "feature" {
		t.Errorf("expected branch feature got %s", newBranch.Name)
	}

	// Check existence
	exists, err := r.Meta.BranchExists("feature")
	if err != nil {
		t.Fatalf("BranchExists failed: %v", err)
	}
	if !exists {
		t.Errorf("expected branch feature to exist")
	}

	// List branches
	branches, err := r.Meta.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("expected 2 branches got %d", len(branches))
	}
}

// Commits
func TestCommitsLifecycle(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	branch := config.DefaultBranch

	commit := &meta.Commit{
		ID:        "abc123",
		Parents:   nil,
		Branch:    branch,
		Message:   "Initial commit",
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: "fileset1",
	}

	id, err := r.Meta.CreateCommit(commit)
	if err != nil {
		t.Fatalf("CreateCommit failed: %v", err)
	}
	if id != commit.ID {
		t.Errorf("expected ID %s got %s", commit.ID, id)
	}

	// Set/Get last commit ID
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

	// Get commit
	c, err := r.Meta.GetCommit(commit.ID)
	if err != nil {
		t.Fatalf("GetCommit failed: %v", err)
	}
	if c.Message != commit.Message {
		t.Errorf("expected message %q got %q", commit.Message, c.Message)
	}

	// AllCommitIDs
	ids, err := r.Meta.AllCommitIDs(branch)
	if err != nil {
		t.Fatalf("AllCommitIDs failed: %v", err)
	}
	if len(ids) != 1 || ids[0] != commit.ID {
		t.Errorf("unexpected AllCommitIDs: %v", ids)
	}

	// GetLastCommitForBranch
	lastCommit, err := r.Meta.GetLastCommitForBranch(branch)
	if err != nil {
		t.Fatalf("GetLastCommitForBranch failed: %v", err)
	}
	if lastCommit.ID != commit.ID {
		t.Errorf("expected last commit ID %s got %s", commit.ID, lastCommit.ID)
	}
}

// HeadRef
func TestHeadRefSetAndGet(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

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

// Storage
func TestRepositoryStorageIntegration(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	if r.Store == nil {
		t.Errorf("expected storage manager to be initialized")
	}
}

// Errors for commit simulation
func TestCommitErrorsSimulation(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	commit := &meta.Commit{
		ID:      "error1",
		Branch:  config.DefaultBranch,
		Message: "test",
	}

	_, err = r.Meta.CreateCommit(commit)
	if err == nil {
		t.Error("expected simulated write error")
	}

	_, err = r.Meta.GetCommit("nonexistent")
	if err == nil {
		t.Error("expected simulated read error")
	}

	err = r.Meta.SetLastCommitID("badbranch", "abc")
	if err == nil {
		t.Error("expected simulated write error")
	}

	_, err = r.Meta.GetLastCommitID("badbranch")
	if err == nil {
		t.Error("expected simulated read error")
	}
}

// Errors for HEAD simulation
func TestHeadErrorsSimulation(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	_, err = r.Meta.GetHeadRef()
	if err == nil {
		t.Error("expected simulated read error for HEAD")
	}

	_, err = r.Meta.SetHeadRef("main")
	if err == nil {
		t.Error("expected simulated write error for HEAD")
	}
}

// AllCommitIDs cycles
func TestAllCommitIDsCycles(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, err := repo.NewRepositoryByPath(tmp)
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

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

	r.Meta.CreateCommit(commitA)
	r.Meta.CreateCommit(commitB)
	r.Meta.SetLastCommitID(config.DefaultBranch, "A")

	ids, err := r.Meta.AllCommitIDs(config.DefaultBranch)
	if err != nil {
		t.Fatalf("AllCommitIDs failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 commits due to cycle guard, got %d", len(ids))
	}
}
