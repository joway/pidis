package db

import (
	"fmt"
	"github.com/joway/pikv/executor"
	"github.com/joway/pikv/types"
	"github.com/joway/pikv/util"
	"github.com/tidwall/redcon"
	"google.golang.org/grpc"
	"net"
)

func ListenRpcServer(database *Database, address string) (*grpc.Server, error) {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	server := NewRpcServer(database)
	go func() {
		if err := server.Serve(listen); err != nil {
			logger.Error("%v", err)
		}
	}()

	return server, nil
}

func ListenRedisProtoServer(database *Database, address string) *redcon.Server {
	sv := redcon.NewServerNetwork(
		"tcp",
		address,
		func(conn redcon.Conn, cmd redcon.Command) {
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
		},
		nil,
		nil,
	)
	go func() {
		if err := sv.ListenAndServe(); err != nil {
			logger.Error("%v", err)
		}
	}()
	return sv
}
