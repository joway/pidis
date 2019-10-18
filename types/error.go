package types

import "errors"

var (
	ErrUnknownCommand = errors.New("ERR unknown command")
	ErrNodeReadOnly   = errors.New("ERR node read only")
	ErrNodeIsMaster   = errors.New("ERR node is master")

	ErrInvalidAOFFormat = errors.New("ERR invalid aof format")

	ErrSyntaxError         = errors.New("ERR syntax error")
	ErrRuntimeError        = errors.New("ERR runtime error")
	ErrInvalidNumberOfArgs = errors.New("ERR invalid number of arguments")

	ErrKeyNotFound = errors.New("ERR key not found")
)
