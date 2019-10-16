package common

import "errors"

const (
	ErrUnknown             = "ERR unknown"
	ErrInvalidNumberOfArgs = "ERR invalid number of arguments"
	ErrSyntaxError         = "ERR syntax error"
	ErrRuntimeError        = "ERR runtime error"
	ErrNodeReadOnly        = "ERR node read only"
)

var (
	ErrNodeIsMaster = errors.New("ERR node is master")
)
