package repotools_test

import (
	"app/internal/config"
	"app/internal/fsio"
	"app/internal/repo"
	"app/internal/repotools"
	"app/internal/storage/block"
	"app/internal/util"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- Helpers ---

func tmpRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bvc-repotools-*")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}

	// Backup globals
	oldCommits := config.CommitsDir
	oldFilesets := config.FilesetsDir
	oldRoot := config.RepoRootOverride

	t.Cleanup(func() {
		config.CommitsDir = oldCommits
		config.FilesetsDir = oldFilesets
		config.RepoRootOverride = oldRoot
		os.RemoveAll(dir)
	})

	// Override repo structure
	config.RepoRootOverride = dir
	config.CommitsDir = filepath.Join(dir, "commits")
	config.FilesetsDir = filepath.Join(dir, "filesets")

	os.MkdirAll(config.CommitsDir, 0o755)
	os.MkdirAll(config.FilesetsDir, 0o755)

	return dir
}

type fakeRepo struct {
	Branches []string
	Err      error
}

func (r *fakeRepo) ListBranches() ([]repo.Branch, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	out := []repo.Branch{}
	for _, n := range r.Branches {
		out = append(out, repo.Branch{Name: n})
	}
	return out, nil
}

func (r *fakeRepo) AllCommitIDs(branch string) ([]string, error) {
	if branch == "badall" {
		return nil, fmt.Errorf("fail allcommit")
	}
	return []string{"c1"}, nil
}

func (r *fakeRepo) GetLastCommitID(branch string) (string, error) {
	if branch == "badlast" {
		return "", fmt.Errorf("fail lastcommit")
	}
	return "c1", nil
}

// --- Tests ---

func TestListAllBlocks_Success(t *testing.T) {
	_ = tmpRepo(t)

	r := &fakeRepo{Branches: []string{"main"}}

	// Write a real repo.Commit JSON
	commit := repo.Commit{
		ID:        "c1",
		Branch:    "main",
		FilesetID: "fs1",
	}
	b, _ := json.Marshal(commit)
	os.WriteFile(filepath.Join(config.CommitsDir, "c1.json"), b, 0o644)

	// Write snapshot.Fileset JSON
	fileset := struct {
		Files []struct {
			Path   string
			Blocks []struct {
				Hash string
				Size int64
			}
		}
	}{
		Files: []struct {
			Path   string
			Blocks []struct {
				Hash string
				Size int64
			}
		}{
			{
				Path: "a.txt",
				Blocks: []struct {
					Hash string
					Size int64
				}{
					{Hash: "h1", Size: 123},
				},
			},
		},
	}
	b, _ = json.Marshal(fileset)
	os.WriteFile(filepath.Join(config.FilesetsDir, "fs1.json"), b, 0o644)

	got, err := repotools.ListAllBlocks(r, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 block, got %d", len(got))
	}
	if _, ok := got["h1"]; !ok {
		t.Errorf("expected h1 in map")
	}
}

func TestListAllBlocks_ErrorBranches(t *testing.T) {
	_ = tmpRepo(t)
	r := &fakeRepo{}

	// ListBranches fails
	r.Err = fmt.Errorf("branchfail")
	if _, err := repotools.ListAllBlocks(r, true); err == nil {
		t.Error("expected error from ListBranches")
	}
	r.Err = nil

	// AllCommitIDs fails
	r.Branches = []string{"badall"}
	if _, err := repotools.ListAllBlocks(r, false); err == nil {
		t.Error("expected error from AllCommitIDs")
	}

	// GetLastCommitID fails
	r.Branches = []string{"badlast"}
	if _, err := repotools.ListAllBlocks(r, true); err == nil {
		t.Error("expected error from GetLastCommitID")
	}
}

func TestCountBlocks_Success(t *testing.T) {
	_ = tmpRepo(t)

	r := &fakeRepo{Branches: []string{"main"}}

	commit := repo.Commit{ID: "c1", Branch: "main", FilesetID: "fs1"}
	b, _ := json.Marshal(commit)
	os.WriteFile(filepath.Join(config.CommitsDir, "c1.json"), b, 0o644)

	fileset := struct {
		Files []struct {
			Path   string
			Blocks []struct {
				Hash string
				Size int64
			}
		}
	}{
		Files: []struct {
			Path   string
			Blocks []struct {
				Hash string
				Size int64
			}
		}{
			{
				Path: "a.txt",
				Blocks: []struct {
					Hash string
					Size int64
				}{
					{Hash: "x", Size: 1},
				},
			},
		},
	}
	b, _ = json.Marshal(fileset)
	os.WriteFile(filepath.Join(config.FilesetsDir, "fs1.json"), b, 0o644)

	n, err := repotools.CountBlocks(r, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestCountBlocks_ErrorCases(t *testing.T) {
	_ = tmpRepo(t)
	r := &fakeRepo{Err: fmt.Errorf("branchfail")}
	if _, err := repotools.CountBlocks(r, true); err == nil {
		t.Error("expected branch error")
	}
}

func TestVerifyBlocksStream(t *testing.T) {
	dir := tmpRepo(t)
	fsio.MkdirAll(dir, 0o755)

	r := &fakeRepo{Branches: []string{"main"}}
	patchReadJSON(t, func(path string, v any) error {
		switch filepath.Base(path) {
		case "c1.json":
			_ = json.Unmarshal(mustJSON(map[string]string{"FilesetID": "fs1"}), v)
		case "fs1.json":
			_ = json.Unmarshal(mustJSON(map[string]any{
				"Files": []map[string]any{
					{"Path": "a.txt", "Blocks": []map[string]any{{"Hash": "x", "Size": 1}}},
				},
			}), v)
		}
		return nil
	})

	out, errCh := repotools.VerifyBlocksStream(r, false)

	var got []block.BlockCheck
	for bc := range out {
		got = append(got, bc)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestVerifyBlocks_MissingRepo(t *testing.T) {
	dir := tmpRepo(t)
	os.RemoveAll(dir)

	r := &fakeRepo{Branches: []string{"main"}}
	err := repotools.VerifyBlocks(r, false)
	if err == nil {
		t.Error("expected missing repo error")
	}
}

// --- Helpers ---
func patchReadJSON(t *testing.T, fn func(string, any) error) {
	t.Helper()
	old := util.ReadJSON
	util.ReadJSON = fn
	t.Cleanup(func() { util.ReadJSON = old })
}

func mustJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
