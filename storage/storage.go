package storage

type Storage interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte, ttl int64) error
	Del(key [][]byte) error
	Close() error
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
