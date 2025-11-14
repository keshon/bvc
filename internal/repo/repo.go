package repo

import (
	"app/internal/config"
	"app/internal/repo/meta"
	"app/internal/repo/store"
	"app/internal/repo/store/snapshot"
)

type Repository struct {
	Config *config.RepoConfig
	Meta   *meta.MetaContext
	Store  *store.StoreContext
}

func NewRepositoryByPath(path string) (*Repository, error) {
	cfg := config.NewRepoConfig(path)
	return NewRepository(cfg)
}

func NewRepository(cfg *config.RepoConfig) (*Repository, error) {
	mt, err := meta.NewMetaDefault(cfg)
	if err != nil {
		return nil, err
	}

	st, err := store.NewStoreDefault(cfg)
	if err != nil {
		return nil, err
	}

	r := &Repository{
		Config: cfg,
		Meta:   mt,
		Store:  st,
	}
	return r, nil
}

func (r *Repository) GetCommitFileset(commitID string) (*snapshot.Fileset, error) {
	commit, err := r.Meta.GetCommit(commitID)
	if err != nil {
		return nil, err
	}
	fs, err := r.Store.Snapshots.Load(commit.FilesetID)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func IsRepoExists(path string) bool {
	cfg := config.NewRepoConfig(path)
	return meta.IsMetaExists(cfg)
}
