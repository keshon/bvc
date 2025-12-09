package stream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"

	"bvc/storage"
)

type Meta struct {
	Name      string   `json:"name"`
	Snapshots []string `json:"snapshots"`
}

type Store struct {
	Backend storage.Storage
	Prefix  string
}

func NewStore(backend storage.Storage, prefix string) *Store {
	return &Store{Backend: backend, Prefix: prefix}
}

func (s *Store) Create(name string) error {
	key := path.Join(s.Prefix, name)
	exists, err := s.Backend.Exists(key)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("stream '%s' exists", name)
	}
	meta := &Meta{Name: name, Snapshots: []string{}}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return s.Backend.Put(key, bytes.NewReader(data))
}

func (s *Store) Load(name string) (*Meta, error) {
	key := path.Join(s.Prefix, name)
	rc, err := s.Backend.Get(key)
	if err != nil {
		return nil, fmt.Errorf("stream '%s' not found", name)
	}
	defer rc.Close()
	var meta Meta
	if err := json.NewDecoder(rc).Decode(&meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *Store) Save(meta *Meta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return s.Backend.Put(path.Join(s.Prefix, meta.Name), bytes.NewReader(data))
}

func (s *Store) List() ([]string, error) {
	keys, err := s.Backend.List(s.Prefix + "/")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, k := range keys {
		names = append(names, k[len(s.Prefix)+1:])
	}
	return names, nil
}

func (s *Store) Delete(name string) error {
	return s.Backend.Delete(path.Join(s.Prefix, name))
}
