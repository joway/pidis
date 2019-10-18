package common

import (
	"context"
	"io"
)

type Database interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte, ttl int64) error
	Del(keys [][]byte) error
	Scan(scanOpts ScanOptions) ([]KVPair, error)
	Exec(args [][]byte) (output []byte, err error)

	SlaveOf(host, port string) error

	Snapshot(ctx context.Context, writer io.Writer) error
	Sync(ctx context.Context, writer io.Writer, offset []byte) error
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
