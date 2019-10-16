package rpc

import (
	"context"
	"github.com/joway/loki"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/rpc/proto"
	"github.com/joway/pikv/util"
)

var logger = loki.New("rpc:server")

type PiKVService struct {
	proto.UnimplementedPiKVServer

	db common.Database
}

func NewPiKVService(db common.Database) *PiKVService {
	return &PiKVService{
		db: db,
	}
}

func (s *PiKVService) Snapshot(req *proto.SnapshotReq, srv proto.PiKV_SnapshotServer) error {
	bus := util.NewStreamBus()
	var err error = nil
	go func() {
		defer bus.Close()
		logger.Info("Fetching snapshot")
		err = s.db.Snapshot(bus)
	}()
	logger.Info("Sending snapshot")
	for buffer := range bus.Read() {
		resp := &proto.SnapshotResp{
			Payload: buffer,
		}
		if err := srv.Send(resp); err != nil {
			return err
		}
	}
	if err != nil {
		logger.Error("Snapshot failed")
	} else {
		logger.Info("Snapshot success")
	}
	return err
}

func (s *PiKVService) Oplog(req *proto.OplogReq, srv proto.PiKV_OplogServer) error {
	bus := util.NewStreamBus()
	ctx := context.TODO()
	var err error = nil
	go func() {
		defer bus.Close()
		logger.Info("Fetching oplog")
		err = s.db.SyncOplog(ctx, bus, req.Offset)
	}()
	logger.Info("Sending oplog")
	for line := range bus.Read() {
		resp := &proto.OplogResp{
			Payload: line,
		}
		if err := srv.Send(resp); err != nil {
			return err
		}
	}
	if err != nil {
		logger.Error("Oplog failed: %v", err)
	} else {
		logger.Info("Oplog success")
	}
	return err
}
