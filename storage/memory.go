package storage

type MemoryStorage struct {
	Storage

	db map[string][]byte
}

func NewMemoryStorage(options Options) (Storage, error) {
	return &MemoryStorage{db: make(map[string][]byte)}, nil
}

func (storage *MemoryStorage) Close() error {
	return nil
}

func (storage *MemoryStorage) Get(key []byte) ([]byte, error) {
	val, ok := storage.db[string(key)]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (storage *MemoryStorage) Set(key, val []byte, ttl int64) error {
	storage.db[string(key)] = val
	return nil
}

func (storage *MemoryStorage) Del(keys [][]byte) error {
	for key := range keys {
		delete(storage.db, string(key))
	}
	return nil
}
