package repo_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/fsio"
	"app/internal/repo"
)

// --- Helper functions --- //
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

// simulate ReadFile/fsio.WriteFile errors to cover error paths
func simulateReadFileError() func() {
	orig := fsio.ReadFile
	fsio.ReadFile = func(_ string) ([]byte, error) {
		return nil, errors.New("simulated read error")
	}
	return func() { fsio.ReadFile = orig }
}

func simulateWriteFileError() func() {
	orig := fsio.WriteFile
	fsio.WriteFile = func(_ string, _ []byte, _ os.FileMode) error {
		return errors.New("simulated write error")
	}
	return func() { fsio.WriteFile = orig }
}

func simulateStatError() func() {
	orig := fsio.StatFile
	fsio.StatFile = func(_ string) (os.FileInfo, error) {
		return nil, errors.New("simulated stat error")
	}
	return func() { fsio.StatFile = orig }
}

// --- Tests --- //

func TestInitAndOpenRepository(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, created, err := repo.InitAt(tmp, "xxh3")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}
	if !created {
		t.Errorf("expected repo to be created")
	}
	if r.Root != tmp {
		t.Errorf("expected Root=%q got %q", tmp, r.Root)
	}

	// Check HEAD file
	headData, err := os.ReadFile(r.HeadFile)
	if err != nil {
		t.Fatalf("failed to read HEAD: %v", err)
	}
	if string(headData) != "ref: branches/main" {
		t.Errorf("unexpected HEAD content: %s", string(headData))
	}

	// Open the same repo
	r2, err := repo.OpenAt(tmp)
	if err != nil {
		t.Fatalf("OpenAt failed: %v", err)
	}
	if r2.Config.HashFormat != "xxh3" {
		t.Errorf("expected hash xxh3 got %s", r2.Config.HashFormat)
	}
}

func TestBranchCreationAndListing(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	// Current branch
	cur, err := r.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if cur.Name != config.DefaultBranch {
		t.Errorf("expected branch %s got %s", config.DefaultBranch, cur.Name)
	}

	// Create new branch
	newBranch, err := r.CreateBranch("feature")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}
	if newBranch.Name != "feature" {
		t.Errorf("expected branch feature got %s", newBranch.Name)
	}

	// Check existence
	exists, err := r.BranchExists("feature")
	if err != nil {
		t.Fatalf("BranchExists failed: %v", err)
	}
	if !exists {
		t.Errorf("expected branch feature to exist")
	}

	// List branches
	branches, err := r.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("expected 2 branches got %d", len(branches))
	}
}

func TestCommitsLifecycle(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	branch := config.DefaultBranch

	commit := &repo.Commit{
		ID:        "abc123",
		Parents:   nil,
		Branch:    branch,
		Message:   "Initial commit",
		Timestamp: time.Now().Format(time.RFC3339),
		FilesetID: "fileset1",
	}

	id, err := r.CreateCommit(commit)
	if err != nil {
		t.Fatalf("CreateCommit failed: %v", err)
	}
	if id != commit.ID {
		t.Errorf("expected ID %s got %s", commit.ID, id)
	}

	// Set/Get last commit ID
	if err := r.SetLastCommitID(branch, commit.ID); err != nil {
		t.Fatalf("SetLastCommitID failed: %v", err)
	}
	lastID, err := r.GetLastCommitID(branch)
	if err != nil {
		t.Fatalf("GetLastCommitID failed: %v", err)
	}
	if lastID != commit.ID {
		t.Errorf("expected lastID %s got %s", commit.ID, lastID)
	}

	// Get commit
	c, err := r.GetCommit(commit.ID)
	if err != nil {
		t.Fatalf("GetCommit failed: %v", err)
	}
	if c.Message != commit.Message {
		t.Errorf("expected message %q got %q", commit.Message, c.Message)
	}

	// AllCommitIDs
	ids, err := r.AllCommitIDs(branch)
	if err != nil {
		t.Fatalf("AllCommitIDs failed: %v", err)
	}
	if len(ids) != 1 || ids[0] != commit.ID {
		t.Errorf("unexpected AllCommitIDs: %v", ids)
	}

	// GetLastCommitForBranch
	lastCommit, err := r.GetLastCommitForBranch(branch)
	if err != nil {
		t.Fatalf("GetLastCommitForBranch failed: %v", err)
	}
	if lastCommit.ID != commit.ID {
		t.Errorf("expected last commit ID %s got %s", commit.ID, lastCommit.ID)
	}
}

func TestHeadRefSetAndGet(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	ref, err := r.SetHeadRef("main")
	if err != nil {
		t.Fatalf("SetHeadRef failed: %v", err)
	}
	if ref.String() != "branches/main" {
		t.Errorf("unexpected HeadRef: %s", ref)
	}

	gotRef, err := r.GetHeadRef()
	if err != nil {
		t.Fatalf("GetHeadRef failed: %v", err)
	}
	if gotRef.String() != ref.String() {
		t.Errorf("HeadRef mismatch: expected %s got %s", ref, gotRef)
	}
}

func TestRepositoryStorageIntegration(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	if r.Storage == nil {
		t.Errorf("expected storage manager to be initialized")
	}
	if r.Storage.Root != r.Root {
		t.Errorf("expected storage.Root=%s got %s", r.Root, r.Storage.Root)
	}
}

func TestInitAtExistingRepo(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	_, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("first InitAt failed: %v", err)
	}

	_, created, err := repo.InitAt(tmp, "")
	if !os.IsExist(err) && !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected os.ErrExist, got %v", err)
	}
	if created {
		t.Errorf("expected created=false for existing repo")
	}
}

func TestOpenAtNonexistentRepo(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	_, err := repo.OpenAt(tmp)
	if err == nil {
		t.Fatal("expected error opening non-existent repo")
	}
}

func TestBranchErrorsSimulation(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	// simulate stat error in BranchExists
	restoreStat := simulateStatError()
	defer restoreStat()
	_, err = r.BranchExists("any")
	if err == nil {
		t.Error("expected simulated stat error")
	}

	// simulate write error in CreateBranch
	restoreWrite := simulateWriteFileError()
	defer restoreWrite()
	_, err = r.CreateBranch("newbranch")
	if err == nil {
		t.Error("expected simulated write error")
	}

	// simulate read error in GetBranch
	restoreRead := simulateReadFileError()
	defer restoreRead()
	_, err = r.GetBranch("nonexistent")
	if err == nil {
		t.Error("expected simulated read error")
	}
}

func TestCommitErrorsSimulation(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	commit := &repo.Commit{
		ID:      "error1",
		Branch:  config.DefaultBranch,
		Message: "test",
	}

	// --- simulate write error on CreateCommit via fsio.CreateTempFile ---
	origCreateTemp := fsio.CreateTempFile
	fsio.CreateTempFile = func(dir, pattern string) (*os.File, error) {
		return nil, errors.New("simulated write error")
	}
	defer func() { fsio.CreateTempFile = origCreateTemp }()

	_, err = r.CreateCommit(commit)
	if err == nil {
		t.Error("expected simulated write error")
	}

	// restore normal CreateTempFile to simulate read error on GetCommit
	origRead := fsio.ReadFile
	fsio.ReadFile = func(_ string) ([]byte, error) {
		return nil, errors.New("simulated read error")
	}
	defer func() { fsio.ReadFile = origRead }()

	_, err = r.GetCommit("nonexistent")
	if err == nil {
		t.Error("expected simulated read error")
	}

	// simulate write error on SetLastCommitID
	origWrite := fsio.WriteFile
	fsio.WriteFile = func(_ string, _ []byte, _ os.FileMode) error {
		return errors.New("simulated write error")
	}
	defer func() { fsio.WriteFile = origWrite }()

	err = r.SetLastCommitID("badbranch", "abc")
	if err == nil {
		t.Error("expected simulated write error")
	}

	// simulate read error on GetLastCommitID
	fsio.ReadFile = func(_ string) ([]byte, error) {
		return nil, errors.New("simulated read error")
	}
	defer func() { fsio.ReadFile = origRead }()

	_, err = r.GetLastCommitID("badbranch")
	if err == nil {
		t.Error("expected simulated read error")
	}
}

func TestHeadErrorsSimulation(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	restoreRead := simulateReadFileError()
	defer restoreRead()
	_, err = r.GetHeadRef()
	if err == nil {
		t.Error("expected simulated read error for HEAD")
	}

	restoreWrite := simulateWriteFileError()
	defer restoreWrite()
	_, err = r.SetHeadRef("main")
	if err == nil {
		t.Error("expected simulated write error for HEAD")
	}
}

func TestAllCommitIDsCycles(t *testing.T) {
	tmp := makeTempDir(t)
	defer os.RemoveAll(tmp)

	r, _, err := repo.InitAt(tmp, "")
	if err != nil {
		t.Fatalf("InitAt failed: %v", err)
	}

	commitA := &repo.Commit{
		ID:        "A",
		Parents:   []string{"B"}, // cycle
		Branch:    config.DefaultBranch,
		Message:   "A",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	commitB := &repo.Commit{
		ID:        "B",
		Parents:   []string{"A"}, // cycle
		Branch:    config.DefaultBranch,
		Message:   "B",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	r.CreateCommit(commitA)
	r.CreateCommit(commitB)
	r.SetLastCommitID(config.DefaultBranch, "A")

	ids, err := r.AllCommitIDs(config.DefaultBranch)
	if err != nil {
		t.Fatalf("AllCommitIDs failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 commits due to cycle guard, got %d", len(ids))
	}
}
