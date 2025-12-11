package storage

import (
	"bytes"
	"io"
)

type MemToLocalStorage struct {
	Memory *MemoryStorage
	Local  *LocalStorage
}

func NewMemToLocalStorage(mem *MemoryStorage, local *LocalStorage) *MemToLocalStorage {
	return &MemToLocalStorage{Memory: mem, Local: local}
}

func (m *MemToLocalStorage) Put(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return err
	}
	if err := m.Memory.Put(key, bytes.NewReader(buf.Bytes())); err != nil {
		return err
	}
	return m.Local.Put(key, bytes.NewReader(buf.Bytes()))
}

func (m *MemToLocalStorage) Get(key string) (io.ReadCloser, error) {
	if exists, _ := m.Memory.Exists(key); exists {
		return m.Memory.Get(key)
	}
	return m.Local.Get(key)
}

func (m *MemToLocalStorage) Delete(key string) error {
	_ = m.Memory.Delete(key)
	return m.Local.Delete(key)
}

func (m *MemToLocalStorage) Exists(key string) (bool, error) {
	if exists, _ := m.Memory.Exists(key); exists {
		return true, nil
	}
	return m.Local.Exists(key)
}

func (m *MemToLocalStorage) List(prefix string) ([]string, error) {
	localKeys, _ := m.Local.List(prefix)
	memKeys, _ := m.Memory.List(prefix)
	all := append(localKeys, memKeys...)
	return all, nil
}
