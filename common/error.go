package common

import "errors"

var (
	ErrUnknown        = errors.New("ERR unknown")
	ErrUnknownCommand = errors.New("ERR unknown command")
	ErrNodeReadOnly   = errors.New("ERR node read only")
	ErrNodeIsMaster   = errors.New("ERR node is master")

	ErrCloseConn = errors.New("ERR close connection")
	ErrShutdown  = errors.New("ERR shutdown node")

	ErrSyntaxError         = errors.New("ERR syntax error")
	ErrRuntimeError        = errors.New("ERR runtime error")
	ErrInvalidNumberOfArgs = errors.New("ERR invalid number of arguments")
)
