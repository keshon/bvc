package engine

// HEAD manager. Keeps current ref, either snapshot or stream.
// Stored as JSON in storage backend.

import (
	"bytes"
	"encoding/json"

	"bvc/pkg/storage"
)

type headMeta struct {
	Mode string `json:"mode"` // "snapshot" or "stream"
	Ref  string `json:"ref"`  // snapshot id or stream name
}

type HeadStore struct {
	Backend storage.Storage
	Key     string // default "HEAD"
}

// NewHeadStore creates a HEAD manager.
func NewHeadStore(backend storage.Storage, key string) *HeadStore {
	if key == "" {
		key = "HEAD"
	}
	return &HeadStore{Backend: backend, Key: key}
}

// Load loads current HEAD.
func (h *HeadStore) Load() (*headMeta, error) {
	rc, err := h.Backend.Get(h.Key)
	if err != nil {
		// no HEAD yet
		return &headMeta{Mode: "", Ref: ""}, nil
	}
	defer rc.Close()

	var m headMeta
	if err := json.NewDecoder(rc).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save writes HEAD.
func (h *HeadStore) Save(mode, ref string) error {
	m := &headMeta{Mode: mode, Ref: ref}
	data, _ := json.MarshalIndent(m, "", "  ")
	return h.Backend.Put(h.Key, bytes.NewReader(data))
}

// Clear removes HEAD completely.
func (h *HeadStore) Clear() error {
	return h.Backend.Delete(h.Key)
}
