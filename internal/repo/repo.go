package repo

import (
	"bvc/internal/blockstore"
	"bvc/internal/engine"
	"bvc/internal/snapshot"
	"bvc/internal/stream"
	"bvc/internal/workfs"
	"bvc/storage"
	"path/filepath"
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
