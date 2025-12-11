package snapshot

// Snapshot store: minimal wrapper over storage.Storage.
// Responsibilities:
// - Save/Load snapshot meta as JSON under key "<prefix>/<id>"
// - List snapshot IDs (trim prefix)
// - Delete snapshot (utility)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bvc/pkg/storage"
	"bvc/pkg/util"
)

type Meta struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	CreatedAt   time.Time           `json:"created_at"`
	Files       map[string][]string `json:"files"` // rel -> []block hashes
}

type Store struct {
	Backend storage.Storage
	Prefix  string
}

func NewStore(backend storage.Storage, prefix string) *Store {
	return &Store{Backend: backend, Prefix: prefix}
}

func (s *Store) Load(id string) (*Meta, error) {
	rc, err := s.Backend.Get(util.JoinNorm(s.Prefix, id))
	if err != nil {
		return nil, fmt.Errorf("snapshot '%s' not found", id)
	}
	defer rc.Close()

	var m Meta
	if err := json.NewDecoder(rc).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) Save(m *Meta) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return s.Backend.Put(util.JoinNorm(s.Prefix, m.ID), bytes.NewReader(data))
}

func (s *Store) List() ([]string, error) {
	keys, err := s.Backend.List(util.Normalize(s.Prefix))
	if err != nil {
		return nil, err
	}
	prefix := util.Normalize(s.Prefix) + "/"
	var ids []string
	for _, k := range keys {
		k = util.Normalize(k)
		if strings.HasPrefix(k, prefix) {
			ids = append(ids, k[len(prefix):])
		}
	}
	return ids, nil
}

func (s *Store) Delete(id string) error {
	return s.Backend.Delete(util.JoinNorm(s.Prefix, id))
}
