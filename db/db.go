package db

import (
	"context"
	"fmt"
	"github.com/joway/pikv/command"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/parser"
	"github.com/joway/pikv/rpc/proto"
	"github.com/joway/pikv/storage"
	"github.com/joway/pikv/types"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"io"
	"os"
	"strings"
)

type Options struct {
	DBDir string
}

type Database struct {
	storage storage.Storage

	following     *Node // slave of
	followingConn *grpc.ClientConn

	aofBuf *AOFBus
}

func NewDatabase(options Options) (*Database, error) {
	dataDir := fmt.Sprintf("%s/data", options.DBDir)
	aofFilePath := fmt.Sprintf("%s/pikv.aof", options.DBDir)
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return nil, err
	}
	storageOpts := storage.Options{
		Storage: storage.TypeBadger,
		Dir:     dataDir,
	}
	store, err := storage.NewStorage(storageOpts)
	if err != nil {
		return nil, err
	}

	//create aofBuf stream
	aofBuf, err := NewAOFBus(aofFilePath)
	if err != nil {
		return nil, err
	}

	database := &Database{
		storage: store,

		aofBuf: aofBuf,
	}

	return database, nil
}

func (db *Database) IsWritable() bool {
	return db.following == nil
}

func (db *Database) Record(cmd [][]byte) error {
	return db.aofBuf.Append(cmd)
}

func (db *Database) Close() error {
	var errs []error
	if err := db.followingConn.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := db.aofBuf.Flush(); err != nil {
		errs = append(errs, err)
	}
	if err := db.aofBuf.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := db.storage.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Errorf("%v", errs)
}

func (db *Database) Exec(context types.Context) ([]byte, types.Action, error) {
	if len(context.Args) == 0 {
		return nil, types.ActionNone, errors.New(common.ErrInvalidNumberOfArgs)
	}
	context.Storage = db.storage
	c := strings.ToUpper(string(context.Args[0]))
	if command.IsWriteCommand(c) {
		if !db.IsWritable() {
			return nil, types.ActionNone, errors.New(common.ErrNodeReadOnly)
		}
		if err := db.Record(context.Args); err != nil {
			return nil, types.ActionNone, err
		}
	}
	out, action := parser.Parse(context)

	return out, action, nil
}

func (db *Database) SlaveOf(host, port string) error {
	db.following = &Node{
		host: host,
		port: port,
	}

	address := fmt.Sprintf("%s:%s", host, port)
	var err error
	db.followingConn, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Following() error {
	if db.following == nil {
		return errors.New(common.ErrNodeIsMaster)
	}

	cli := proto.NewPiKVClient(db.followingConn)
	ctx := context.TODO()
	offset := xid.New().Bytes()
	//fetch snapshot
	snapStream, err := cli.Snapshot(ctx, &proto.SnapshotReq{})
	if err != nil {
		return err
	}
	for {
		resp, err := snapStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		resp.GetPayload()
	}
	//fetch and replay oplog
	oplogStream, err := cli.Oplog(ctx, &proto.OplogReq{
		Offset: offset,
	})
	if err != nil {
		return err
	}
	for {
		resp, err := oplogStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		resp.GetPayload()
		//replay oplog
	}

	return nil
}

func (db *Database) SyncOplog(context context.Context, writer io.Writer, offset []byte) error {
	return db.aofBuf.Sync(context, writer, offset)
}

func (db *Database) Snapshot(writer io.Writer) error {
	if err := db.storage.Snapshot(writer); err != nil {
		return err
	}
	return nil
}

func (db *Database) LoadSnapshot(reader io.Reader) error {
	return db.storage.LoadSnapshot(reader)
}
