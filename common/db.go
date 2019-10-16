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
	Snapshot(writer io.Writer) error
	SyncOplog(context context.Context, writer io.Writer, offset []byte) error
}
