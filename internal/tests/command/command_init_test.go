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

// Helper to create a temporary directory and switch into it
func tmpWorkDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bvc-init-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	oldDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.Chdir(oldDir)
		os.RemoveAll(dir)
	})

	return dir
}

// Run the init command with args in the current working directory
func runInit(t *testing.T, args ...string) error {
	t.Helper()
	cmd := &initcmd.Command{}
	ctx := &command.Context{Args: args}
	return cmd.Run(ctx)
}

// Check if a repository was created
func checkRepoExists(t *testing.T, path string) *repo.Repository {
	t.Helper()
	r, err := repo.OpenAt(path)
	if err != nil {
		t.Fatalf("expected repository at %q, got error: %v", path, err)
	}
	return r
}

// --- Tests ---

func TestInit_DefaultRepo(t *testing.T) {
	tmpWorkDir(t)
	if err := runInit(t); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	r := checkRepoExists(t, ".bvc")
	if r.Config.HashFormat != config.DefaultHash {
		t.Errorf("expected hash %q, got %q", config.DefaultHash, r.Config.HashFormat)
	}
}

func TestInit_CustomInitialBranch(t *testing.T) {
	tmpWorkDir(t)
	args := []string{"--initial-branch", "dev"}
	if err := runInit(t, args...); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	r := checkRepoExists(t, ".bvc")
	head, _ := r.GetHeadRef()
	if !strings.HasSuffix(head.String(), "dev") {
		t.Errorf("expected HEAD to point to dev branch, got %q", head)
	}
}

func TestInit_BareRepo(t *testing.T) {
	tmpWorkDir(t)
	args := []string{"--bare"}
	if err := runInit(t, args...); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	r := checkRepoExists(t, filepath.Join(".", config.RepoDir))
	if r == nil {
		t.Fatal("bare repository not created")
	}
}

func TestInit_SeparateDir(t *testing.T) {
	dir := tmpWorkDir(t)
	sep := filepath.Join(dir, "myrepo")
	args := []string{"--separate-bvc-dir", sep}
	if err := runInit(t, args...); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Pointer file created in working dir
	data, err := fsio.ReadFile(config.RepoPointerFile)
	if err != nil {
		t.Fatalf("expected pointer file in working dir, got error: %v", err)
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
	dir := tmpWorkDir(t)
	repoDir := filepath.Join(dir, config.RepoDir)
	_, _, _ = repo.InitAt(repoDir, config.DefaultHash)

	// Re-init should succeed silently in quiet mode
	args := []string{"--quiet"}
	if err := runInit(t, args...); err != nil {
		t.Fatalf("re-init failed: %v", err)
	}
}

func TestInit_QuietMode(t *testing.T) {
	tmpWorkDir(t)
	args := []string{"--quiet"}
	if err := runInit(t, args...); err != nil {
		t.Fatalf("init failed in quiet mode: %v", err)
	}
}
