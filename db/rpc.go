package db

import (
	"context"
	"github.com/joway/loki"
	"github.com/joway/pikv/proto"
	"github.com/joway/pikv/util"
	"google.golang.org/grpc"
)

var rpcLogger = loki.New("pikv:db:rpc")

func NewRpcServer(database *Database) *grpc.Server {
	server := grpc.NewServer()
	proto.RegisterPiKVServer(server, NewPiKVService(database))

	return server
}

type PiKVService struct {
	proto.UnimplementedPiKVServer

	db *Database
}

func NewPiKVService(database *Database) *PiKVService {
	return &PiKVService{
		db: database,
	}
}

func (s *PiKVService) Snapshot(req *proto.SnapshotReq, srv proto.PiKV_SnapshotServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := util.NewStreamBus()
	var err error = nil
	go func() {
		defer bus.Close()
		rpcLogger.Info("fetching snapshot")
		err = s.db.storage.Snapshot(ctx, bus)
	}()
	rpcLogger.Info("sending snapshot")
	for buffer := range bus.Read() {
		resp := &proto.SnapshotResp{
			Payload: buffer,
		}
		if err := srv.Send(resp); err != nil {
			return err
		}
	}
	if err != nil {
		rpcLogger.Error("snapshot failed")
	} else {
		rpcLogger.Info("snapshot success")
	}
	return err
}

func (s *PiKVService) Oplog(req *proto.OplogReq, srv proto.PiKV_OplogServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var syncErr = make(chan error, 1)
	bus := util.NewStreamBus()
	go func() {
		defer bus.Close()
		rpcLogger.Info("fetching oplog")
		syncErr <- s.db.Sync(ctx, bus, req.Offset)
	}()
	rpcLogger.Info("sending oplog")
	for {
		select {
		case e := <-syncErr:
			if e != nil {
				rpcLogger.Error("oplog sync failed: %v", e)
				return e
			}
			rpcLogger.Error("oplog sync finished")
			return nil
		case payload := <-bus.Read():
			resp := &proto.OplogResp{
				Payload: payload,
			}
			if err := srv.Send(resp); err != nil {
				return err
			}
		}
	}
}
