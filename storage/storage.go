package storage

import "io"

type Storage interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte, ttl int64) error
	Del(keys [][]byte) error
	Close() error
	Snapshot(writer io.Writer) error
	LoadSnapshot(reader io.Reader) error
}

const (
	TypeBadger = "badger"
	TypeMemory = "memory"
)

type Options struct {
	Storage string
	Dir     string
}

func NewStorage(options Options) (Storage, error) {
	switch options.Storage {
	default:
		return NewBadgerStorage(options)
	case TypeBadger:
		return NewBadgerStorage(options)
	case TypeMemory:
		return NewMemoryStorage(options)
	}
}
