package repo

import (
	"path/filepath"

	"github.com/keshon/bvc/internal/blockstore"
	"github.com/keshon/bvc/internal/engine"
	"github.com/keshon/bvc/internal/snapshot"
	"github.com/keshon/bvc/internal/stream"
	"github.com/keshon/bvc/internal/workfs"
	"github.com/keshon/bvc/pkg/storage"
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
