package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"app/internal/config"
)

// InitRepo initializes the repository and returns its name and error
func InitRepo() (string, error) {
	if exists, err := repoExists(); err != nil {
		return "", fmt.Errorf("failed to check if repo exists: %w", err)
	} else if exists {
		return "", errors.New("repository already exists")
	}

	// Ensure all necessary directories exist
	dirs := []string{
		config.RepoDir,
		config.CommitsDir,
		config.FilesetsDir,
		config.BranchesDir,
		config.ObjectsDir,
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return "", fmt.Errorf("failed to create directory %q: %w", d, err)
		}
	}

	// Ensure the main branch exists
	mainHead := filepath.Join(config.BranchesDir, config.DefaultBranch)
	if _, err := os.Stat(mainHead); os.IsNotExist(err) {
		if err := os.WriteFile(mainHead, []byte(""), 0o644); err != nil {
			return "", fmt.Errorf("failed to create default branch file %q: %w", mainHead, err)
		}

		headPath := filepath.Join(config.RepoDir, config.HeadFile)
		if err := os.WriteFile(headPath, []byte("ref: branches/"+config.DefaultBranch), 0o644); err != nil {
			return "", fmt.Errorf("failed to initialize HEAD file %q: %w", headPath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to stat default branch file %q: %w", mainHead, err)
	}

	// Determine repository name
	repoDir := config.RepoDir
	repoName := filepath.Base(repoDir)

	return repoName, nil
}

func repoExists() (bool, error) {
	_, err := os.Stat(config.RepoDir)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat repo: %w", err)
}
