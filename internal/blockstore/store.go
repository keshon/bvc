package blockstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/keshon/bvc/pkg/storage"
)

// Store — thin wrapper around storage.Storage that exposes content-addressable API.
type Store struct {
	Backend storage.Storage
	Prefix  string
}

// NewStore returns blockstore wrapper.
func NewStore(b storage.Storage, prefix string) *Store { return &Store{Backend: b, Prefix: prefix} }

// PutIfAbsent reads from r, computes hash, stores under key if missing, and returns hash.
func (s *Store) PutIfAbsent(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	h := hashBytes(data)
	key := s.fullKey(h)

	exists, err := s.Backend.Exists(key)
	if err != nil {
		return "", err
	}
	if exists {
		// hash collision check
		rc, err := s.Backend.Get(key)
		if err == nil {
			defer rc.Close()
			existing, _ := io.ReadAll(rc)
			if !bytes.Equal(existing, data) {
				return "", fmt.Errorf("hash collision for %s", h)
			}
		}
		return h, nil
	}
	if err := s.Backend.Put(key, bytes.NewReader(data)); err != nil {
		return "", err
	}
	return h, nil
}

// Get copies block to writer; verifies hash if verify==true (cheap double-check).
func (s *Store) Get(hash string, w io.Writer, verify bool) error {
	key := s.fullKey(hash)
	rc, err := s.Backend.Get(key)
	if err != nil {
		return err
	}
	defer rc.Close()
	if !verify {
		_, err = io.Copy(w, rc)
		return err
	}
	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	if hashBytes(data) != hash {
		return fmt.Errorf("integrity check failed for %s", hash)
	}
	_, err = w.Write(data)
	return err
}

// ------------------------ Helpers ------------------------

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (s *Store) fullKey(hash string) string {
	if s.Prefix == "" {
		return hash
	}
	return s.Prefix + "/" + hash
}
