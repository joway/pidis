package parser

import (
	"github.com/joway/pikv/command"
	"github.com/joway/pikv/types"
	"strings"
)

type Action int

const (
	Close Action = iota
	Shutdown
)

func Parse(context types.Context) ([]byte, Action) {
	var action Action
	switch strings.ToUpper(string(context.Args[0])) {
	default:
		return Unknown(context), action
	case command.QUIT:
		return Nil(context), Close
	case command.SHUTDOWN:
		return Nil(context), Shutdown
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
