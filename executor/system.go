package executor

import (
	"fmt"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/util"
)

type SystemExecutor struct {
	BaseExecutor
}

func (e SystemExecutor) Exec(db common.Database, args [][]byte) ([]byte, Action) {
	switch e.cmd {
	default:
		return e.Unknown(db, args)
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
	}
}

func (e SystemExecutor) Unknown(db common.Database, args [][]byte) ([]byte, Action) {
	return util.MessageError(fmt.Sprintf(
		"ERR unknown command '%s'",
		e.cmd,
	)), ActionNone
}

func (SystemExecutor) Ping(db common.Database, args [][]byte) ([]byte, Action) {
	return util.MessageString("PONG"), ActionNone
}

func (SystemExecutor) Echo(db common.Database, args [][]byte) ([]byte, Action) {
	if len(args) == 2 {
		return util.Message(args[1]), ActionNone
	} else {
		return nil, ActionInvalidNumberOfArgs
	}
}

func (SystemExecutor) Quit(db common.Database, args [][]byte) ([]byte, Action) {
	return util.MessageOK(), ActionClose
}

func (SystemExecutor) Shutdown(db common.Database, args [][]byte) ([]byte, Action) {
	return util.MessageOK(), ActionShutdown
}

func (e SystemExecutor) SlaveOf(db common.Database, args [][]byte) ([]byte, Action) {
	host := string(args[1])
	port := string(args[2])
	if err := db.SlaveOf(host, port); err != nil {
		logger.Error("%v", err)
		return util.MessageError(fmt.Sprintf("%s Failed", e.cmd)), ActionNone
	}
	return util.MessageOK(), ActionNone
}
