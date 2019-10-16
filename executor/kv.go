package executor

import (
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/util"
	"strconv"
)

type KVExecutor struct {
	BaseExecutor
}

func (e KVExecutor) Exec(db common.Database, args [][]byte) ([]byte, error) {
	switch e.cmd {
	case GET:
		return e.Get(db, args)
	case DEL:
		return e.Del(db, args)
	case SET:
		return e.Set(db, args)
	default:
		return nil, common.ErrUnknownCommand
	}
}

func (KVExecutor) Get(db common.Database, args [][]byte) ([]byte, error) {
	if len(args) != 2 {
		return nil, common.ErrInvalidNumberOfArgs
	}
	key := args[1]
	val, err := db.Get(key)
	if err != nil {
		return util.MessageNull(), nil
	}
	return util.Message(val), nil
}

func (e KVExecutor) Set(db common.Database, args [][]byte) ([]byte, error) {
	if !(len(args) >= 3 && len(args) <= 4) {
		return nil, common.ErrInvalidNumberOfArgs
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
			return nil, common.ErrSyntaxError
		}
	}

	if err := db.Set(key, val, ttl); err != nil {
		logger.Error("%v", err)
		return nil, common.ErrRuntimeError
	}
	return util.MessageOK(), nil
}

func (e KVExecutor) Del(db common.Database, args [][]byte) ([]byte, error) {
	if len(args) < 2 {
		return nil, common.ErrInvalidNumberOfArgs
	}
	if err := db.Del(args[1:]); err != nil {
		return nil, common.ErrSyntaxError
	}

	return util.MessageInt(int64(len(args) - 1)), nil
}
