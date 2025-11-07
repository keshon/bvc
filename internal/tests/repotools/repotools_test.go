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

// --- helpers ---

// tmpRepo creates a temporary repo and returns (dir, cfg)
func tmpRepo(t *testing.T) (string, *config.RepoConfig) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.NewRepoConfig(dir)
	return dir, cfg
}

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

// --- Fake repo for testing ---
type fakeRepo struct {
	Branches []string
}

func (r *fakeRepo) ListBranches() ([]repo.Branch, error) {
	if r.Branches == nil {
		return nil, fmt.Errorf("ListBranches failed")
	}
	var out []repo.Branch
	for _, name := range r.Branches {
		out = append(out, repo.Branch{Name: name})
	}
	return out, nil
}

func (r *fakeRepo) AllCommitIDs(branch string) ([]string, error) {
	switch branch {
	case "badall":
		return nil, fmt.Errorf("AllCommitIDs failed")
	case "badlast":
		return []string{"cX"}, nil // Ensure GetLastCommitID is called
	default:
		return []string{"c1"}, nil
	}
}

func (r *fakeRepo) GetLastCommitID(branch string) (string, error) {
	if branch == "badlast" {
		return "", fmt.Errorf("GetLastCommitID failed")
	}
	return "c1", nil
}

// --- Tests ---

func TestListAllBlocks_Success(t *testing.T) {
	_, cfg := tmpRepo(t)

	r := &fakeRepo{Branches: []string{"main"}}

	// Write a commit JSON
	commit := repo.Commit{
		ID:        "c1",
		Branch:    "main",
		FilesetID: "fs1",
	}
	b, _ := json.Marshal(commit)
	os.MkdirAll(cfg.CommitsDir(), 0o755)
	os.WriteFile(filepath.Join(cfg.CommitsDir(), "c1.json"), b, 0o644)

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
	os.MkdirAll(cfg.FilesetsDir(), 0o755)
	os.WriteFile(filepath.Join(cfg.FilesetsDir(), "fs1.json"), b, 0o644)

	got, err := repotools.ListAllBlocks(r, cfg, true)
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
	_, cfg := tmpRepo(t)

	// --- ListBranches fails ---
	r := &fakeRepo{Branches: nil} // triggers ListBranches error
	if _, err := repotools.ListAllBlocks(r, cfg, true); err == nil {
		t.Error("expected error from ListBranches")
	}

	// --- AllCommitIDs fails ---
	r = &fakeRepo{Branches: []string{"badall"}}
	if _, err := repotools.ListAllBlocks(r, cfg, false); err == nil {
		t.Error("expected error from AllCommitIDs")
	}

	// --- GetLastCommitID fails ---
	r = &fakeRepo{Branches: []string{"badlast"}}

	// ensure commits dir exists
	if err := os.MkdirAll(cfg.CommitsDir(), 0o755); err != nil {
		t.Fatalf("failed to create commits dir: %v", err)
	}

	// write dummy commit for branch "badlast"
	dummyCommit := map[string]string{"ID": "c1", "Branch": "badlast", "FilesetID": "fs1"}
	b, _ := json.Marshal(dummyCommit)
	os.WriteFile(filepath.Join(cfg.CommitsDir(), "c1.json"), b, 0o644)

	// now ListAllBlocks will call GetLastCommitID and trigger the error
	if _, err := repotools.ListAllBlocks(r, cfg, true); err == nil {
		t.Error("expected error from GetLastCommitID")
	}

}

func TestCountBlocks_Success(t *testing.T) {
	_, cfg := tmpRepo(t)
	r := &fakeRepo{Branches: []string{"main"}}

	commit := repo.Commit{ID: "c1", Branch: "main", FilesetID: "fs1"}
	b, _ := json.Marshal(commit)
	os.MkdirAll(cfg.CommitsDir(), 0o755)
	os.WriteFile(filepath.Join(cfg.CommitsDir(), "c1.json"), b, 0o644)

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
	os.MkdirAll(cfg.FilesetsDir(), 0o755)
	os.WriteFile(filepath.Join(cfg.FilesetsDir(), "fs1.json"), b, 0o644)

	n, err := repotools.CountBlocks(r, cfg, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestCountBlocks_ErrorCases(t *testing.T) {
	_, cfg := tmpRepo(t)

	// fakeRepo that fails on ListBranches
	r := &fakeRepo{
		Branches: nil, // nil triggers ListBranches error
	}

	if _, err := repotools.CountBlocks(r, cfg, true); err == nil {
		t.Error("expected branch error")
	}
}

func TestVerifyBlocksStream(t *testing.T) {
	dir, cfg := tmpRepo(t)
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

	out, errCh := repotools.VerifyBlocksStream(r, cfg, false)

	var got []block.BlockCheck
	for bc := range out {
		got = append(got, bc)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestVerifyBlocks_MissingRepo(t *testing.T) {
	dir, cfg := tmpRepo(t)
	os.RemoveAll(dir)

	r := &fakeRepo{Branches: []string{"main"}}
	err := repotools.VerifyBlocks(r, cfg, false)
	if err == nil {
		t.Error("expected missing repo error")
	}
}
