package storage

import (
	"github.com/patrickmn/go-cache"
	"time"
)

type MemoryStorage struct {
	Storage

	db *cache.Cache
}

func NewMemoryStorage(options Options) (Storage, error) {
	db := cache.New(cache.NoExpiration, 1*time.Millisecond)
	return &MemoryStorage{db: db}, nil
}

func (storage *MemoryStorage) Close() error {
	return nil
}

func (storage *MemoryStorage) Get(key []byte) ([]byte, error) {
	val, ok := storage.db.Get(string(key))
	if !ok {
		return nil, nil
	}
	return val.([]byte), nil
}

func (storage *MemoryStorage) Set(key, val []byte, ttl int64) error {
	storage.db.Set(string(key), val, time.Duration(ttl)*time.Millisecond)
	return nil
}

func (storage *MemoryStorage) Del(keys [][]byte) error {
	for key := range keys {
		storage.db.Delete(string(key))
	}
	return nil
}
