package core

import (
	"app/internal/config"
	"os"
	"path/filepath"
)

// InitRepo initializes the repository
func InitRepo() error {
	for _, d := range []string{config.RepoDir, config.CommitsDir, config.FilesetsDir, config.BranchesDir, config.ObjectsDir} {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		}
	}

	mainHead := filepath.Join(config.BranchesDir, config.DefaultBranch)
	if _, err := os.Stat(mainHead); os.IsNotExist(err) {
		if err := os.WriteFile(mainHead, []byte(""), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(config.RepoDir, config.HeadFile), []byte("ref: branches/"+config.DefaultBranch), 0644); err != nil {
			return err
		}
	}
	return nil
}
