package core

import (
	"app/internal/config"
	"app/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zeebo/xxh3"
)

type Commit struct {
	ID        string   `json:"id"`
	Parents   []string `json:"parents"`
	Branch    string   `json:"branch"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	FilesetID string   `json:"fileset_id"`
}

func LoadCommit(commitID string) (*Commit, error) {
	var c Commit
	path := filepath.Join(config.CommitsDir, commitID+".json")
	if err := util.ReadJSON(path, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func LastCommit(branch string) (string, error) {
	path := filepath.Join(config.BranchesDir, branch)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func SetLastCommit(branch, commitID string) error {
	path := filepath.Join(config.BranchesDir, branch)
	return os.WriteFile(path, []byte(commitID), 0644)
}

// not used
func GenerateCommitID(filesetID string, parents ...string) string {
	// concatenate input
	data := []byte(filesetID)
	for _, p := range parents {
		data = append(data, []byte(p)...)
	}
	data = append(data, []byte(time.Now().String())...)

	hash := xxh3.Hash128(data)

	// convert to bytes
	hashBytes := hash.Bytes()
	return fmt.Sprintf("%x", hashBytes[:8])
}
