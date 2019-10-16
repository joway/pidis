package executor

import (
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/util"
	"strconv"
)

type KVExecutor struct {
	BaseExecutor
}

func (KVExecutor) Get(db common.Database, args [][]byte) ([]byte, Action) {
	if len(args) != 2 {
		return util.MessageError(common.ErrInvalidNumberOfArgs), ActionNone
	}
	key := args[1]
	val, err := db.Get(key)
	if err != nil {
		return util.MessageNull(), ActionNone
	}
	return util.Message(val), ActionNone
}

func (e KVExecutor) Set(db common.Database, args [][]byte) ([]byte, Action) {
	if !(len(args) >= 3 && len(args) <= 4) {
		return nil, ActionInvalidNumberOfArgs
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
			return nil, ActionInvalidSyntax
		}
	}

	if err := db.Set(key, val, ttl); err != nil {
		logger.Error("%v", err)
		return nil, ActionRuntimeError
	}
	return util.MessageOK(), ActionNone
}

func (e KVExecutor) Del(db common.Database, args [][]byte) ([]byte, Action) {
	if len(args) < 2 {
		return nil, ActionInvalidNumberOfArgs
	}
	if err := db.Del(args[1:]); err != nil {
		return nil, ActionInvalidSyntax
	}

	return util.MessageInt(int64(len(args) - 1)), ActionNone
}

func (e KVExecutor) Exec(db common.Database, args [][]byte) ([]byte, Action) {
	switch e.cmd {
	case GET:
		return e.Get(db, args)
	case DEL:
		return e.Del(db, args)
	case SET:
		return e.Set(db, args)
	default:
		return nil, ActionUnknown
	}
}
