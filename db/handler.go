package db

import (
	"fmt"
	"github.com/joway/pidis/executor"
	"github.com/joway/pidis/types"
	"github.com/joway/pidis/util"
	"github.com/tidwall/redcon"
)

func GetRedisCmdHandler(database *Database) func(conn redcon.Conn, cmd redcon.Command) {
	return func(conn redcon.Conn, cmd redcon.Command) {
		defer func() {
			if err := recover(); err != nil {
				conn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
			}
		}()
		result, err := database.Exec(cmd.Args)
		//handle action
		if result != nil {
			switch result.Action() {
			case executor.ActionSlaveOf:
				host := cmd.Args[1]
				port := cmd.Args[2]
				if err := database.SlaveOf(string(host), string(port)); err != nil {
					logger.Error("ERR slaveof: %v", err)
					conn.WriteRaw(util.MessageError(err.Error()))
					return
				}
			case executor.ActionConnClose:
				if err := conn.Close(); err != nil {
					logger.Error("connection close Failed:\n%v", err)
				}
			case executor.ActionShutdown:
				logger.Fatal("shutting server down, bye bye")
			}
		}
		//handle err
		switch err {
		case nil:
			conn.WriteRaw(result.Output())
		case types.ErrUnknownCommand:
			conn.WriteRaw(util.MessageError(fmt.Sprintf(
				"ERR unknown command '%s'",
				cmd.Args[0],
			)))
		case types.ErrRuntimeError:
			conn.WriteRaw(util.MessageError(types.ErrRuntimeError.Error()))
		case types.ErrInvalidNumberOfArgs:
			conn.WriteRaw(util.MessageError(types.ErrInvalidNumberOfArgs.Error()))
		case types.ErrSyntaxError:
			conn.WriteRaw(util.MessageError(types.ErrSyntaxError.Error()))
		default:
			logger.Error("Unhandled error : %v", err)
			conn.WriteRaw(util.MessageError(err.Error()))
		}
	}

}
