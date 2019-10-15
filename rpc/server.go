package rpc

import (
	"github.com/joway/pikv/db"
	"github.com/joway/pikv/rpc/proto"
	"google.golang.org/grpc"
)

func NewRpcServer(database *db.Database) *grpc.Server {
	server := grpc.NewServer()
	proto.RegisterOplogServer(server, NewOplogService(database))

	return server
}
