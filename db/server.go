package db

import (
	"fmt"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/rpc"
	"github.com/joway/pikv/util"
	"github.com/tidwall/redcon"
	"google.golang.org/grpc"
	"net"
)

func ListenRpcServer(database common.Database, address string) (*grpc.Server, error) {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	server := rpc.NewRpcServer(database)
	go func() {
		if err := server.Serve(listen); err != nil {
			logger.Error("%v", err)
		}
	}()

	return server, nil
}

func ListenRedisProtoServer(database common.Database, address string) *redcon.Server {
	sv := redcon.NewServerNetwork(
		"tcp",
		address,
		func(conn redcon.Conn, cmd redcon.Command) {
			defer func() {
				if err := recover(); err != nil {
					conn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
				}
			}()
			out, err := database.Exec(cmd.Args)
			switch err {
			case nil:
				conn.WriteRaw(out)
			case common.ErrUnknownCommand:
				conn.WriteRaw(util.MessageError(fmt.Sprintf(
					"ERR unknown command '%s'",
					cmd.Args[0],
				)))
			case common.ErrUnknown:
				conn.WriteRaw(util.MessageError(common.ErrUnknown.Error()))
			case common.ErrRuntimeError:
				conn.WriteRaw(util.MessageError(common.ErrRuntimeError.Error()))
			case common.ErrInvalidNumberOfArgs:
				conn.WriteRaw(util.MessageError(common.ErrInvalidNumberOfArgs.Error()))
			case common.ErrSyntaxError:
				conn.WriteRaw(util.MessageError(common.ErrSyntaxError.Error()))
			case common.ErrCloseConn:
				if err := conn.Close(); err != nil {
					logger.Error("connection close Failed:\n%v", err)
				}
			case common.ErrShutdown:
				logger.Fatal("shutting server down, bye bye")
			default:
				logger.Error("Unhandled error : %v", err)
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
