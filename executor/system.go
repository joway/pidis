package executor

import (
	"github.com/joway/pidis/storage"
	"github.com/joway/pidis/types"
	"github.com/joway/pidis/util"
)

type SystemExecutor struct {
	BaseExecutor
}

func (e SystemExecutor) Exec(store storage.Storage, args [][]byte) (*Result, error) {
	switch e.cmd {
	case PING:
		return e.Ping(store, args)
	case ECHO:
		return e.Echo(store, args)
	case QUIT:
		return e.Quit(store, args)
	case SHUTDOWN:
		return e.Shutdown(store, args)
	case SLAVEOF:
		return e.SlaveOf(store, args)
	default:
		return nil, types.ErrUnknownCommand
	}
}

func (SystemExecutor) Ping(store storage.Storage, args [][]byte) (*Result, error) {
	return &Result{output: util.MessageString("PONG")}, nil
}

func (SystemExecutor) Echo(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) == 2 {
		return &Result{output: util.Message(args[1])}, nil
	} else {
		return nil, types.ErrInvalidNumberOfArgs
	}
}

func (SystemExecutor) Quit(store storage.Storage, args [][]byte) (*Result, error) {
	return &Result{
		output: util.MessageOK(),
		action: ActionConnClose,
	}, nil
}

func (SystemExecutor) Shutdown(store storage.Storage, args [][]byte) (*Result, error) {
	return &Result{
		output: util.MessageOK(),
		action: ActionShutdown,
	}, nil
}

func (e SystemExecutor) SlaveOf(store storage.Storage, args [][]byte) (*Result, error) {
	if len(args) != 3 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	return &Result{output: util.MessageOK(), action: ActionSlaveOf}, nil
}
