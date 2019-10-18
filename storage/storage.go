package storage

import (
	"context"
	"io"
)

type Storage interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte, ttl int64) error
	Del(keys [][]byte) error
	Scan(scanOpts ScanOptions) ([]KVPair, error)
	Close() error

	Snapshot(ctx context.Context, writer io.Writer) error
	LoadSnapshot(ctx context.Context, reader io.Reader) error
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

type ScanOptions struct {
	Pattern      string
	Limit        int
	IncludeValue bool
}

type KVPair struct {
	Key, Val []byte
}

func (p *KVPair) SetKey(key []byte) {
	p.Key = key
}

func (p *KVPair) SetVal(val []byte) {
	p.Val = val
}
