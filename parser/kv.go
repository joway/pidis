package parser

import (
	"github.com/joway/loki"
	"github.com/joway/pikv/types"
	"github.com/tidwall/redcon"
	"strconv"
)

func Get(context types.Context) []byte {
	args := context.Args
	if len(args) != 2 {
		return redcon.AppendError(context.Out, ErrInvalidNumberOfArgs)
	}
	key := args[1]
	val, err := context.Storage.Get(key)
	if err != nil {
		return redcon.AppendNull(context.Out)
	}
	return redcon.AppendBulk(context.Out, val)
}

func Set(context types.Context) []byte {
	args := context.Args
	if !(len(args) >= 3 && len(args) <= 4) {
		return redcon.AppendError(context.Out, ErrInvalidNumberOfArgs)
	}

	key := args[1]
	val := args[2]
	var ttl int64 = 0
	var err error
	if len(args) == 4 {
		ttl, err = strconv.ParseInt(string(args[3]), 10, 64)
		if err != nil {
			return redcon.AppendError(context.Out, ErrSyntaxError)
		}
	}

	if err := context.Storage.Set(key, val, ttl); err != nil {
		loki.Error("%v", err)
		return redcon.AppendError(context.Out, ErrRuntimeError)
	}
	return redcon.AppendOK(context.Out)
}

func Del(context types.Context) []byte {
	args := context.Args
	if len(args) < 2 {
		return redcon.AppendError(context.Out, ErrInvalidNumberOfArgs)
	}
	if err := context.Storage.Del(args[1:]); err != nil {
		return redcon.AppendError(context.Out, ErrSyntaxError)
	}

	return redcon.AppendInt(context.Out, int64(len(args)-1))
}
