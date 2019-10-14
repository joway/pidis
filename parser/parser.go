package parser

import (
	"github.com/joway/pikv/command"
	"github.com/joway/pikv/types"
	"strings"
)

func Parse(context types.Context) ([]byte, types.Action) {
	var action types.Action
	switch strings.ToUpper(string(context.Args[0])) {
	default:
		return Unknown(context), action
	case command.QUIT:
		return Nil(context), types.Close
	case command.SHUTDOWN:
		return Nil(context), types.Shutdown
	case command.PING:
		return Ping(context), action
	case command.ECHO:
		return Echo(context), action
	case command.SET:
		return Set(context), action
	case command.GET:
		return Get(context), action
	case command.DEL:
		return Del(context), action
	}
}
