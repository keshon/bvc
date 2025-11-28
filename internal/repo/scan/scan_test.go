package scan_test

import (
	"path/filepath"
	"testing"

	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/fs"
	"github.com/keshon/bvc/internal/repo/meta"
	"github.com/keshon/bvc/internal/repo/scan"
)

// minimal mock for MetaContextInterface
type mockMeta struct {
	lastCommitFiles []*meta.Commit
	indexFiles      []string
}

func (m *mockMeta) GetCurrentBranch() (*meta.Branch, error) {
	return &meta.Branch{Name: "main"}, nil
}
func (m *mockMeta) GetBranch(name string) (meta.Branch, error) {
	return meta.Branch{Name: name}, nil
}
func (m *mockMeta) ListBranches() ([]meta.Branch, error) { return nil, nil }
func (m *mockMeta) CreateBranch(name string) (meta.Branch, error) {
	return meta.Branch{Name: name}, nil
}
func (m *mockMeta) BranchExists(name string) (bool, error) { return false, nil }

func (m *mockMeta) GetCommit(commitID string) (*meta.Commit, error)  { return nil, nil }
func (m *mockMeta) CreateCommit(commit *meta.Commit) (string, error) { return "", nil }
func (m *mockMeta) SetLastCommitID(branch, commitID string) error    { return nil }
func (m *mockMeta) GetLastCommitID(branch string) (string, error)    { return "", nil }
func (m *mockMeta) AllCommitIDs(branch string) ([]string, error)     { return nil, nil }
func (m *mockMeta) GetCommitsForBranch(branch string) ([]*meta.Commit, error) {
	return m.lastCommitFiles, nil
}
func (m *mockMeta) GetLastCommitForBranch(branch string) (*meta.Commit, error) {
	if len(m.lastCommitFiles) > 0 {
		return m.lastCommitFiles[len(m.lastCommitFiles)-1], nil
	}
	return nil, nil
}

func (m *mockMeta) GetHeadRef() (meta.HeadRef, error) {
	return meta.HeadRef(""), nil
}

func (m *mockMeta) SetHeadRef(branch string) (meta.HeadRef, error) {
	return meta.HeadRef(""), nil
}

func (m *mockMeta) GetConfig() *config.RepoConfig {
	return &config.RepoConfig{RepoDir: ".bvc"}
}

// helper to write snapshot JSON to memoryFS
func writeSnapshot(fsys fs.FS, files []string, id string) error {
	_ = fsys.MkdirAll(".bvc/snapshots", 0o755)
	data := []byte(`{"files":[`)
	for i, f := range files {
		if i > 0 {
			data = append(data, ',')
		}
		data = append(data, []byte(`{"path":"`+filepath.ToSlash(f)+`"}`)...)
	}
	data = append(data, ']', '}')
	return fsys.WriteFile(".bvc/snapshots/"+id+".json", data, 0o644)
}

// --- Example test using this mock ---

func TestScannerTrackedAndStaged(t *testing.T) {
	memFS := fs.NewMemoryFS()
	headFiles := []string{"a.txt"}
	commit := &meta.Commit{FilesetID: "HEAD"}
	mock := &mockMeta{lastCommitFiles: []*meta.Commit{commit}}

	// write snapshot JSON
	if err := writeSnapshot(memFS, headFiles, "HEAD"); err != nil {
		t.Fatal(err)
	}

	// staged file
	memFS.MkdirAll(".bvc", 0o755)
	memFS.WriteFile(".bvc/index.json", []byte(`[{"path":"b.txt"}]`), 0o644)

	// working tree files
	memFS.WriteFile("a.txt", []byte("a"), 0o644)
	memFS.WriteFile("b.txt", []byte("b"), 0o644)
	memFS.WriteFile("c.txt", []byte("c"), 0o644)

	scanner := scan.NewScanner(".", mock, memFS)
	tracked, staged, ignored, err := scanner.ScanAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(tracked) != 1 || tracked[0] != "a.txt" {
		t.Errorf("tracked = %v, want [a.txt]", tracked)
	}
	if len(staged) != 1 || staged[0] != "b.txt" {
		t.Errorf("staged = %v, want [b.txt]", staged)
	}
	if len(ignored) != 0 {
		t.Errorf("ignored = %v, want []", ignored)
	}
}
