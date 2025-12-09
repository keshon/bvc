package storage

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
	}
}

func (m *MemoryStorage) Put(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = buf.Bytes()
	return nil
}

func (m *MemoryStorage) Get(key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (m *MemoryStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MemoryStorage) Exists(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

func (m *MemoryStorage) List(prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []string
	for k := range m.data {
		if prefix == "" || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			result = append(result, k)
		}
	}
	return result, nil
}
