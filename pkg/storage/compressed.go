package storage

import (
	"compress/gzip"
	"io"
)

type CompressedStorage struct {
	Inner Storage
}

func NewCompressedStorage(inner Storage) *CompressedStorage {
	return &CompressedStorage{Inner: inner}
}

func (c *CompressedStorage) Put(key string, r io.Reader) error {
	pr, pw := io.Pipe()
	go func() {
		gw := gzip.NewWriter(pw)
		_, err := io.Copy(gw, r)
		gw.Close()
		pw.CloseWithError(err)
	}()
	return c.Inner.Put(key, pr)
}

func (c *CompressedStorage) Get(key string) (io.ReadCloser, error) {
	rc, err := c.Inner.Get(key)
	if err != nil {
		return nil, err
	}
	gr, err := gzip.NewReader(rc)
	if err != nil {
		rc.Close()
		return nil, err
	}
	return struct {
		io.Reader
		io.Closer
	}{gr, rc}, nil
}

func (c *CompressedStorage) Delete(key string) error {
	return c.Inner.Delete(key)
}

func (c *CompressedStorage) Exists(key string) (bool, error) {
	return c.Inner.Exists(key)
}

func (c *CompressedStorage) List(prefix string) ([]string, error) {
	return c.Inner.List(prefix)
}
