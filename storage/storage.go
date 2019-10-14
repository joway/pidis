package storage

type Storage interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte, ttl int64) error
	Del(key [][]byte) error
	Close() error
}

type Options struct {
	Dir string
}
