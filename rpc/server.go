package rpc

import (
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/rpc/proto"
	"google.golang.org/grpc"
)

func NewRpcServer(database common.Database) *grpc.Server {
	server := grpc.NewServer()
	proto.RegisterPiKVServer(server, NewPiKVService(database))

	return server
}
