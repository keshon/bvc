package util

import (
	"bvc/internal/blockstore"
	"bvc/internal/engine"
	"bvc/internal/snapshot"
	"bvc/internal/stream"
	"bvc/internal/workfs"
	"bvc/storage"
	"io"
	"path/filepath"
)

func CopyClose(dst io.Writer, src io.ReadCloser) error {
	defer src.Close()
	_, err := io.Copy(dst, src)
	return err
}

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
