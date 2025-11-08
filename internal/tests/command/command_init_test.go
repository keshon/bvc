package init_test

import (
	"app/internal/command"
	initcmd "app/internal/command/init"
	"app/internal/config"
	"app/internal/fsio"
	"app/internal/repo"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helpers
func makeTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// Run the init command with args in the current working directory
func runInitAt(t *testing.T, workDir string, args ...string) error {
	t.Helper()
	old, _ := os.Getwd()
	defer os.Chdir(old)

	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	cmd := &initcmd.Command{}
	ctx := &command.Context{Args: args}
	return cmd.Run(ctx)
}

// Check if a repository was created
func checkRepoExists(t *testing.T, repoPath string) *repo.Repository {
	t.Helper()
	r, err := repo.NewRepositoryByPath(repoPath)
	if err != nil {
		t.Fatalf("expected repository at %q, got error: %v", repoPath, err)
	}
	return r
}

// --- Tests ---

func TestInit_CustomInitialBranch(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, ".bvc")

	args := []string{"--initial-branch", "dev"}
	if err := runInitAt(t, dir, args...); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	r := checkRepoExists(t, repoDir)
	head, _ := r.Meta.GetHeadRef()
	if !strings.HasSuffix(head.String(), "dev") {
		t.Errorf("expected HEAD to point to dev branch, got %q", head)
	}
}
func TestInit_SeparateDir(t *testing.T) {
	dir := t.TempDir()
	sep := filepath.Join(dir, "myrepo")

	if err := runInitAt(t, dir, "--separate-bvc-dir", sep); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// pointer file
	pointerFile := filepath.Join(dir, config.RepoPointerFile)
	data, err := fsio.ReadFile(pointerFile)
	if err != nil {
		t.Fatalf("expected pointer file, got error: %v", err)
	}
	if string(data) != sep {
		t.Errorf("expected pointer file content %q, got %q", sep, string(data))
	}

	r := checkRepoExists(t, sep)
	if r == nil {
		t.Fatal("separate repository not created")
	}
}

func TestInit_ReinitExistingRepo(t *testing.T) {
	dir := makeTempDir(t)
	repoDir := filepath.Join(dir, config.RepoDir)
	_, _ = repo.NewRepositoryByPath(repoDir)

	// Re-init should succeed silently in quiet mode
	args := []string{"--quiet"}
	if err := runInitAt(t, dir, args...); err != nil {
		t.Fatalf("re-init failed: %v", err)
	}
}

func TestInit_QuietMode(t *testing.T) {
	dir := makeTempDir(t)
	args := []string{"--quiet"}
	if err := runInitAt(t, dir, args...); err != nil {
		t.Fatalf("init failed in quiet mode: %v", err)
	}
}

func TestInit_RespectsPointerFile(t *testing.T) {
	dir := makeTempDir(t)
	sep := filepath.Join(dir, "myrepo")

	// init with separate dir
	if err := runInitAt(t, dir, "--separate-bvc-dir", sep); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// now re-init without --separate-bvc-dir
	if err := runInitAt(t, dir); err != nil {
		t.Fatalf("re-init failed: %v", err)
	}

	r := checkRepoExists(t, sep)
	if r == nil {
		t.Fatal("repo not found at pointer location after re-init")
	}
}

func TestInit_IgnoredInitialBranch(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, ".bvc")

	// init default
	if err := runInitAt(t, dir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// re-init with custom branch
	args := []string{"--initial-branch", "dev"}
	err := runInitAt(t, dir, args...)
	if err != nil {
		t.Fatalf("re-init failed: %v", err)
	}

	// HEAD should still be default
	r := checkRepoExists(t, repoDir)
	head, _ := r.Meta.GetHeadRef()
	if !strings.HasSuffix(head.String(), config.DefaultBranch) {
		t.Errorf("HEAD branch changed unexpectedly: got %s", head)
	}
}
