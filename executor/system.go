package executor

import (
	"fmt"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/util"
)

type SystemExecutor struct {
	BaseExecutor
}

func (e SystemExecutor) Exec(db common.Database, args [][]byte) ([]byte, error) {
	switch e.cmd {
	case PING:
		return e.Ping(db, args)
	case ECHO:
		return e.Echo(db, args)
	case QUIT:
		return e.Quit(db, args)
	case SHUTDOWN:
		return e.Shutdown(db, args)
	case SLAVEOF:
		return e.SlaveOf(db, args)
	default:
		return nil, common.ErrUnknownCommand
	}
}

func (SystemExecutor) Ping(db common.Database, args [][]byte) ([]byte, error) {
	return util.MessageString("PONG"), nil
}

func (SystemExecutor) Echo(db common.Database, args [][]byte) ([]byte, error) {
	if len(args) == 2 {
		return util.Message(args[1]), nil
	} else {
		return nil, common.ErrInvalidNumberOfArgs
	}
}

func (SystemExecutor) Quit(db common.Database, args [][]byte) ([]byte, error) {
	return util.MessageOK(), common.ErrCloseConn
}

func (SystemExecutor) Shutdown(db common.Database, args [][]byte) ([]byte, error) {
	return util.MessageOK(), common.ErrShutdown
}

func (e SystemExecutor) SlaveOf(db common.Database, args [][]byte) ([]byte, error) {
	host := string(args[1])
	port := string(args[2])
	if err := db.SlaveOf(host, port); err != nil {
		logger.Error("%v", err)
		return util.MessageError(fmt.Sprintf("%s Failed", e.cmd)), nil
	}
	return util.MessageOK(), nil
}
