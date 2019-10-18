package executor

import (
	"github.com/joway/pikv/storage"
	"github.com/joway/pikv/types"
	"github.com/joway/pikv/util"
	"strconv"
)

type KVExecutor struct {
	BaseExecutor
}

func (e KVExecutor) Exec(store storage.Storage, args [][]byte) (*Result, error) {
	switch e.cmd {
	case GET:
		return e.Get(store, args)
	case DEL:
		return e.Del(store, args)
	case SET:
		return e.Set(store, args)
	case KEYS:
		return e.Keys(store, args)
	default:
		return nil, types.ErrUnknownCommand
	}
}

func (KVExecutor) Get(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) != 2 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	key := args[1]
	val, err := store.Get(key)
	if err == types.ErrKeyNotFound {
		return &Result{output: util.MessageNull()}, nil
	}
	if err != nil {
		return &Result{output: util.MessageError(err.Error())}, nil
	}
	return &Result{output: util.Message(val)}, nil
}

func (e KVExecutor) Set(store storage.Storage, args [][]byte) (*Result, error) {
	if !(len(args) >= 3 && len(args) <= 4) {
		return nil, types.ErrInvalidNumberOfArgs
	}

	var (
		key       = args[1]
		val       = args[2]
		ttl int64 = 0
		err error
	)

	if len(args) == 4 {
		ttl, err = strconv.ParseInt(string(args[3]), 10, 64)
		if err != nil {
			return nil, types.ErrSyntaxError
		}
	}

	if err := store.Set(key, val, ttl); err != nil {
		logger.Error("%v", err)
		return nil, types.ErrRuntimeError
	}
	return &Result{output: util.MessageOK()}, nil
}

func (e KVExecutor) Del(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) < 2 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	if err := store.Del(args[1:]); err != nil {
		return nil, types.ErrSyntaxError
	}

	return &Result{output: util.MessageInt(int64(len(args) - 1))}, nil
}

func (e KVExecutor) Keys(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) < 2 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	pattern := args[1]
	pairs, err := store.Scan(storage.ScanOptions{Pattern: string(pattern)})
	if err != nil {
		return nil, err
	}
	var keys [][]byte
	for _, p := range pairs {
		keys = append(keys, p.Key)
	}
	return &Result{output: util.MessageArray(keys)}, nil
}
