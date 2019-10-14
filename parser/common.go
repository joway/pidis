package parser

import (
	"fmt"
	"github.com/joway/pikv/types"
	"github.com/tidwall/redcon"
)

func Nil(context types.Context) []byte {
	return redcon.AppendOK(context.Out)
}

func Unknown(context types.Context) []byte {
	return redcon.AppendError(
		context.Out,
		fmt.Sprintf(
			"ERR unknown command '%s'",
			string(context.Args[0]),
		),
	)
}

func Ping(context types.Context) []byte {
	args := context.Args
	if len(args) == 1 {
		return redcon.AppendString(context.Out, "PONG")
	} else if len(args) == 2 {
		return redcon.AppendBulk(context.Out, args[1])
	} else {
		return redcon.AppendError(context.Out, ErrInvalidNumberOfArgs)
	}
}

func Echo(context types.Context) []byte {
	args := context.Args
	if len(args) == 2 {
		return redcon.AppendBulk(context.Out, args[1])
	} else {
		return redcon.AppendError(context.Out, ErrInvalidNumberOfArgs)
	}
}
