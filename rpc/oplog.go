package rpc

import (
	"context"
	"github.com/joway/pikv/db"
	"github.com/joway/pikv/rpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OplogService struct {
	proto.UnimplementedOplogServer

	db *db.Database
}

func NewOplogService(db *db.Database) *OplogService {
	return &OplogService{
		db: db,
	}
}

func (s *OplogService) Sync(ctx context.Context, req *proto.SyncReq) (*proto.SyncResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sync not implemented")
}
