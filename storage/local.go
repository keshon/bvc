package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	BasePath string
}

func NewLocalStorage(base string) *LocalStorage {
	return &LocalStorage{BasePath: base}
}

func (l *LocalStorage) path(key string) string {
	return filepath.Join(l.BasePath, key)
}

func (l *LocalStorage) Put(key string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(l.path(key)), 0755); err != nil {
		return err
	}
	f, err := os.Create(l.path(key))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (l *LocalStorage) Get(key string) (io.ReadCloser, error) {
	return os.Open(l.path(key))
}

func (l *LocalStorage) Delete(key string) error {
	return os.Remove(l.path(key))
}

func (l *LocalStorage) Exists(key string) (bool, error) {
	_, err := os.Stat(l.path(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LocalStorage) List(prefix string) ([]string, error) {
	var result []string
	err := filepath.Walk(l.BasePath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(l.BasePath, p)
		rel = filepath.ToSlash(rel)
		pfx := strings.TrimSuffix(prefix, "/") + "/"
		if pfx == "/" {
			pfx = ""
		}
		if pfx == "" || strings.HasPrefix(rel, pfx) {
			result = append(result, rel)
		}
		return nil
	})
	return result, err
}

func (l *LocalStorage) PutBatch(items map[string]io.Reader) error {
	for key, r := range items {
		if err := l.Put(key, r); err != nil {
			return err
		}
	}
	return nil
}

func (l *LocalStorage) DeleteBatch(keys []string) error {
	for _, key := range keys {
		if err := l.Delete(key); err != nil {
			return err
		}
	}
	return nil
}
