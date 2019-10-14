package storage

import "sync"

type Map struct {
	sync.RWMutex
	store map[string][]byte
}

type MemoryStorage struct {
	Storage

	db Map
}

func NewMemoryStorage(options Options) (Storage, error) {
	return &MemoryStorage{db: Map{
		store: map[string][]byte{},
	}}, nil
}

func (storage *MemoryStorage) Close() error {
	return nil
}

func (storage *MemoryStorage) Get(key []byte) ([]byte, error) {
	storage.db.RLock()
	val, ok := storage.db.store[string(key)]
	storage.db.RUnlock()
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (storage *MemoryStorage) Set(key, val []byte, ttl int64) error {
	storage.db.Lock()
	storage.db.store[string(key)] = val
	storage.db.Unlock()
	return nil
}

func (storage *MemoryStorage) Del(keys [][]byte) error {
	for key := range keys {
		storage.db.Lock()
		delete(storage.db.store, string(key))
		storage.db.Unlock()
	}
	return nil
}
