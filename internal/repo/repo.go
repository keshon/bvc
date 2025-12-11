package repo

import (
	"path/filepath"

	"bvc/internal/blockstore"
	"bvc/internal/engine"
	"bvc/internal/snapshot"
	"bvc/internal/stream"
	"bvc/internal/workfs"
	"bvc/pkg/storage"
)

func OpenRepo(path string) (*engine.Engine, error) {
	repoDir := filepath.Join(path, workfs.DefaultRepoDir)
	store := storage.NewLocalStorage(repoDir)

	blocks := blockstore.NewStore(store, "blocks")
	snaps := snapshot.NewStore(store, "snaps")
	streams := stream.NewStore(store, "streams")
	ignore := workfs.LoadIgnore(path)

	repo := engine.NewEngine(path, blocks, snaps, streams, ignore, store)

	return repo, nil
}
