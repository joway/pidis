package db

import (
	"context"
	"github.com/joway/loki"
	"github.com/joway/pidis/proto"
	"github.com/joway/pidis/util"
	"google.golang.org/grpc"
)

var rpcLogger = loki.New("pidis:db:rpc")

func NewRpcServer(database *Database) *grpc.Server {
	server := grpc.NewServer()
	proto.RegisterPidisServer(server, NewPidisService(database))

	return server
}

type PidisService struct {
	proto.UnimplementedPidisServer

	db *Database
}

func NewPidisService(database *Database) *PidisService {
	return &PidisService{
		db: database,
	}
}

func (s *PidisService) Snapshot(req *proto.SnapshotReq, srv proto.Pidis_SnapshotServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := util.NewStreamBus(1)
	var err error = nil
	go func() {
		defer bus.Close()
		rpcLogger.Info("fetching snapshot")
		err = s.db.storage.Snapshot(ctx, bus)
	}()
	rpcLogger.Info("sending snapshot")
	payloadMaxBytes := 1024 * 1024 // 1MB
	for buffer := range bus.Read() {
		for len(buffer) > 0 {
			var payload []byte
			if len(buffer) >= payloadMaxBytes {
				payload = buffer[:payloadMaxBytes]
				buffer = buffer[payloadMaxBytes:]
			} else {
				payload = buffer
				buffer = []byte{}
			}
			resp := &proto.SnapshotResp{
				Payload: payload,
			}
			if err := srv.Send(resp); err != nil {
				return err
			}
		}
	}
	if err != nil {
		rpcLogger.Error("snapshot failed")
	} else {
		rpcLogger.Info("snapshot success")
	}
	return err
}

func (s *PidisService) Oplog(req *proto.OplogReq, srv proto.Pidis_OplogServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var syncErr = make(chan error, 1)
	bus := util.NewStreamBus(1024)
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
