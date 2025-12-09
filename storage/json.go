package storage

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type JSONStorage struct {
	BasePath string
}

func NewJSONStorage(base string) *JSONStorage {
	return &JSONStorage{BasePath: base}
}

func (j *JSONStorage) Put(key string, r io.Reader) error {
	var obj interface{}
	if err := json.NewDecoder(r).Decode(&obj); err != nil {
		return err
	}
	path := filepath.Join(j.BasePath, key+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(obj)
}

func (j *JSONStorage) Get(key string) (io.ReadCloser, error) {
	path := filepath.Join(j.BasePath, key+".json")
	return os.Open(path)
}

func (j *JSONStorage) Delete(key string) error {
	return os.Remove(filepath.Join(j.BasePath, key+".json"))
}

func (j *JSONStorage) Exists(key string) (bool, error) {
	_, err := os.Stat(filepath.Join(j.BasePath, key+".json"))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (j *JSONStorage) List(prefix string) ([]string, error) {
	var result []string
	err := filepath.Walk(j.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(j.BasePath, path)
			result = append(result, rel[:len(rel)-5]) // remove ".json"
		}
		return nil
	})
	return result, err
}
