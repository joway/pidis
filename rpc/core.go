package rpc

import (
	"context"
	"github.com/joway/pikv/db"
	"github.com/joway/pikv/rpc/proto"
	"github.com/joway/pikv/util"
)

type PiKVService struct {
	proto.UnimplementedPiKVServer

	db *db.Database
}

func NewPiKVService(db *db.Database) *PiKVService {
	return &PiKVService{
		db: db,
	}
}

func (s *PiKVService) Snapshot(req *proto.SnapshotReq, srv proto.PiKV_SnapshotServer) error {
	bus := util.NewStreamBus()
	var err error = nil
	go func() {
		defer bus.Close()
		err = s.db.Snapshot(bus)
	}()
	for buffer := range bus.Read() {
		resp := &proto.SnapshotResp{
			Payload: buffer,
		}
		if err := srv.Send(resp); err != nil {
			return err
		}
	}
	return err
}

func (s *PiKVService) Oplog(req *proto.OplogReq, srv proto.PiKV_OplogServer) error {
	bus := util.NewStreamBus()
	ctx := context.TODO()
	var err error = nil
	go func() {
		defer bus.Close()
		err = s.db.SyncOplog(ctx, bus, req.Offset)
	}()
	for line := range bus.Read() {
		resp := &proto.OplogResp{
			Payload: line,
		}
		if err := srv.Send(resp); err != nil {
			return err
		}
	}
	return err
}
