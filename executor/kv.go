package executor

import (
	"github.com/joway/pidis/storage"
	"github.com/joway/pidis/types"
	"github.com/joway/pidis/util"
	"strconv"
	"strings"
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
	case SETNX:
		return e.Set(store, append(args, []byte("NX")))
	case KEYS:
		return e.Keys(store, args)
	case EXISTS:
		return e.Exists(store, args)
	case INCR:
		return e.Incr(store, args)
	case TTL:
		return e.TTL(store, args)
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
	if !(len(args) >= 3 || len(args) <= 6) {
		return nil, types.ErrInvalidNumberOfArgs
	}

	var (
		key           = args[1]
		val           = args[2]
		ttl    uint64 = 0
		retErr error
	)
	if len(args) >= 5 {
		switch strings.ToUpper(string(args[3])) {
		case "EX":
			ttl, retErr = strconv.ParseUint(string(args[4]), 10, 64)
			ttl *= 1000
			if retErr != nil {
				return nil, types.ErrSyntaxError
			}
		case "PX":
			ttl, retErr = strconv.ParseUint(string(args[4]), 10, 64)
			if retErr != nil {
				return nil, types.ErrSyntaxError
			}
		}
	}
	setMode := ""
	if len(args) == 4 || len(args) == 6 {
		setMode = strings.ToUpper(string(args[len(args)-1]))
	}
	switch setMode {
	case "NX":
		//TODO: performance, use IsExisted check
		_, err := store.Get(key)
		if err == types.ErrKeyNotFound {
			retErr = store.Set(key, val, ttl)
		} else {
			return &Result{output: util.MessageNull()}, nil
		}
	case "XX":
		//TODO: performance, use IsExisted check
		_, err := store.Get(key)
		if err == types.ErrKeyNotFound {
			return &Result{output: util.MessageNull()}, nil
		}
		retErr = store.Set(key, val, ttl)
	default:
		retErr = store.Set(key, val, ttl)
	}
	if retErr != nil {
		logger.Error("%v", retErr)
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

func (e KVExecutor) TTL(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) < 2 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	key := args[1]
	ttl, err := store.TTL(key)
	code := int64(ttl) / 1000
	if err == types.ErrKeyNotFound {
		code = -2
	} else if err != nil {
		return nil, err
	} else if ttl == 0 {
		code = -1
	}
	return &Result{output: util.MessageInt(code)}, nil
}

func (e KVExecutor) Exists(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) < 2 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	keys := args[1:]
	var count int64 = 0
	for _, key := range keys {
		//TODO: performance tuning, use key only get
		_, err := store.Get(key)
		if err == nil {
			count++
		}
	}
	return &Result{output: util.MessageInt(count)}, nil
}

func (e KVExecutor) Incr(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) != 2 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	key := args[1]
	var num int64 = 0
	val, err := store.Get(key)
	if err == types.ErrKeyNotFound {
		num = 1
		val = []byte("1")
	} else if err == nil {
		num, err = strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			return &Result{output: util.MessageError(err.Error())}, nil
		}
		num++
		val = []byte(strconv.FormatInt(num, 10))
	} else {
		return nil, err
	}
	if err := store.Set(key, val, 0); err != nil {
		return nil, err
	}
	return &Result{output: util.MessageInt(num)}, nil
}
