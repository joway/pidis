package common

import (
	"context"
	"io"
)

type Database interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte, ttl int64) error
	Del(keys [][]byte) error

	SlaveOf(host, port string) error

	Snapshot(ctx context.Context, writer io.Writer) error
	Sync(ctx context.Context, writer io.Writer, offset []byte) error
}
