package storage

import "io"

type Storage interface {
	Put(key string, r io.Reader) error
	Get(key string) (io.ReadCloser, error)
	Delete(key string) error
	Exists(key string) (bool, error)
	List(prefix string) ([]string, error)
	PutBatch(items map[string]io.Reader) error
	DeleteBatch(keys []string) error
}
