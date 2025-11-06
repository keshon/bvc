package repotools_test

import (
	"app/internal/config"
	"app/internal/fsio"
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
	// Override repo dirs
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

func (r *fakeRepo) ListBranches() ([]struct{ Name string }, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	out := []struct{ Name string }{}
	for _, n := range r.Branches {
		out = append(out, struct{ Name string }{Name: n})
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

var repoOpenAt func(string) (interface{}, error)

// Override util.ReadJSON temporarily
func patchReadJSON(t *testing.T, fn func(string, any) error) {
	t.Helper()
	old := util.ReadJSON
	util.ReadJSON = fn
	t.Cleanup(func() { util.ReadJSON = old })
}

// --- Tests ---

func TestListAllBlocks_Success(t *testing.T) {
	dir := tmpRepo(t)
	defer os.RemoveAll(dir)

	// Fake repo with one branch
	r := &fakeRepo{Branches: []string{"main"}}
	repoOpenAt = func(string) (interface{}, error) { return r, nil }

	// Create fake commit and fileset data
	commitPath := filepath.Join(config.CommitsDir, "c1.json")
	filesetPath := filepath.Join(config.FilesetsDir, "fs1.json")

	commitData := map[string]string{"FilesetID": "fs1"}
	fsData := map[string]any{"Files": []map[string]any{
		{"Path": "a.txt", "Blocks": []map[string]any{{"Hash": "h1", "Size": 123}}},
	}}

	os.WriteFile(commitPath, mustJSON(commitData), 0o644)
	os.WriteFile(filesetPath, mustJSON(fsData), 0o644)

	// Real util.ReadJSON works fine
	got, err := repotools.ListAllBlocks(false)
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
	dir := tmpRepo(t)
	defer os.RemoveAll(dir)

	// --- Case: repo.OpenAt fails
	repoOpenAt = func(string) (interface{}, error) { return nil, fmt.Errorf("openfail") }
	if _, err := repotools.ListAllBlocks(false); err == nil {
		t.Error("expected error from repo.OpenAt")
	}

	// --- Case: ListBranches fails
	repoOpenAt = func(string) (interface{}, error) {
		return &fakeRepo{Err: fmt.Errorf("branchfail")}, nil
	}
	if _, err := repotools.ListAllBlocks(false); err == nil {
		t.Error("expected error from ListBranches")
	}

	// --- Case: AllCommitIDs fails
	repoOpenAt = func(string) (interface{}, error) {
		return &fakeRepo{Branches: []string{"badall"}}, nil
	}
	if _, err := repotools.ListAllBlocks(true); err == nil {
		t.Error("expected error from AllCommitIDs")
	}

	// --- Case: GetLastCommitID fails
	repoOpenAt = func(string) (interface{}, error) {
		return &fakeRepo{Branches: []string{"badlast"}}, nil
	}
	if _, err := repotools.ListAllBlocks(false); err == nil {
		t.Error("expected error from GetLastCommitID")
	}
}

func TestCountBlocks_Success(t *testing.T) {
	dir := tmpRepo(t)
	defer os.RemoveAll(dir)

	r := &fakeRepo{Branches: []string{"main"}}
	repoOpenAt = func(string) (interface{}, error) { return r, nil }

	os.WriteFile(filepath.Join(config.CommitsDir, "c1.json"),
		mustJSON(map[string]string{"FilesetID": "fs1"}), 0o644)
	os.WriteFile(filepath.Join(config.FilesetsDir, "fs1.json"),
		mustJSON(map[string]any{"Files": []map[string]any{
			{"Blocks": []map[string]any{{"Hash": "x"}}},
		}}), 0o644)

	n, err := repotools.CountBlocks(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestCountBlocks_ErrorCases(t *testing.T) {
	dir := tmpRepo(t)
	defer os.RemoveAll(dir)

	// repo.OpenAt fails
	repoOpenAt = func(string) (interface{}, error) { return nil, fmt.Errorf("no repo") }
	if _, err := repotools.CountBlocks(false); err == nil {
		t.Error("expected error on open")
	}

	// ListBranches fails
	repoOpenAt = func(string) (interface{}, error) {
		return &fakeRepo{Err: fmt.Errorf("branchfail")}, nil
	}
	if _, err := repotools.CountBlocks(false); err == nil {
		t.Error("expected branch error")
	}
}

func TestVerifyBlocksStream(t *testing.T) {
	dir := tmpRepo(t)
	defer os.RemoveAll(dir)

	// Make sure root exists
	fsio.MkdirAll(config.ResolveRepoRoot(), 0o755)

	// Patch dependencies
	repoOpenAt = func(string) (interface{}, error) { return &fakeRepo{Branches: []string{"main"}}, nil }
	patchReadJSON(t, func(path string, v any) error {
		switch {
		case filepath.Base(path) == "c1.json":
			_ = json.Unmarshal(mustJSON(map[string]string{"FilesetID": "fs1"}), v)
		case filepath.Base(path) == "fs1.json":
			_ = json.Unmarshal(mustJSON(map[string]any{
				"Files": []map[string]any{
					{"Path": "a.txt", "Blocks": []map[string]any{{"Hash": "x", "Size": 1}}},
				},
			}), v)
		}
		return nil
	})

	out, errCh := repotools.VerifyBlocksStream(false)

	var got []block.BlockCheck
	for bc := range out {
		got = append(got, bc)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestVerifyBlocksStream_ErrorRepo(t *testing.T) {
	dir := tmpRepo(t)
	defer os.RemoveAll(dir)

	repoOpenAt = func(string) (interface{}, error) { return nil, fmt.Errorf("bad repo") }

	out, errCh := repotools.VerifyBlocksStream(false)
	select {
	case <-out:
	default:
	}
	if err := <-errCh; err == nil {
		t.Error("expected repo error")
	}
}

func TestVerifyBlocks_MissingRepo(t *testing.T) {
	dir := tmpRepo(t)
	defer os.RemoveAll(dir)

	// Ensure root missing
	os.RemoveAll(config.ResolveRepoRoot())

	err := repotools.VerifyBlocks(false)
	if err == nil {
		t.Error("expected missing repo error")
	}
}

// --- Helpers ---
func mustJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
