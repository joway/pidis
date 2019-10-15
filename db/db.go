package db

import (
	"context"
	"fmt"
	"github.com/joway/loki"
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
	"path"
	"strings"
)

type Options struct {
	DBDir string
}

type Database struct {
	dir     string
	storage storage.Storage

	//signals
	sigFollowing chan bool

	following     *Node // slave of
	followingConn *grpc.ClientConn

	aofBuf *AOFBus
}

func NewDatabase(options Options) (*Database, error) {
	dataDir := path.Join(options.DBDir, "data")
	aofFilePath := fmt.Sprintf(options.DBDir, "pikv.aof")
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
		dir:     options.DBDir,
		storage: store,

		aofBuf: aofBuf,
	}

	return database, nil
}

func (db *Database) Run() {
	go func() {
		if err := db.Daemon(); err != nil {
			loki.Fatal("%v", err)
		}
	}()
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

func (db *Database) Daemon() error {
	var followingContext context.Context
	for {
		select {
		case sig := <-db.sigFollowing:
			if sig {
				go func() {
					defer func() {
						db.sigFollowing <- false
					}()
					followingContext := context.TODO()
					if err := db.Following(followingContext); err != nil {
						loki.Fatal("%v", err)
					}
				}()
			} else {
				followingContext.Done()
			}
		}
	}
}

func (db *Database) SlaveOf(host, port string) error {
	address := fmt.Sprintf("%s:%s", host, port)
	var err error
	db.followingConn, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	db.following = &Node{
		host: host,
		port: port,
	}
	db.sigFollowing <- true
	return nil
}

func (db *Database) Following(context context.Context) error {
	if db.following == nil {
		return errors.New(common.ErrNodeIsMaster)
	}

	cli := proto.NewPiKVClient(db.followingConn)
	offset := xid.New().Bytes()
	//fetch snapshot
	snapStream, err := cli.Snapshot(context, &proto.SnapshotReq{})
	snapPath := path.Join(db.dir, "data")
	if err != nil {
		return err
	}
	snapFile, err := os.OpenFile(snapPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	//save snapshot
	for {
		resp, err := snapStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		content := resp.GetPayload()
		if _, err := snapFile.Write(content); err != nil {
			return err
		}
	}
	//restore snapshot
	if err := db.LoadSnapshot(snapFile); err != nil {
		return err
	}
	//fetch and replay oplog
	oplogStream, err := cli.Oplog(context, &proto.OplogReq{
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

		line := resp.GetPayload()
		//replay oplog
		//TODO: concurrent
		_, args := AOFDecode(line)
		_, _, err = db.Exec(types.Context{
			Args: args,
		})
		if err != nil {
			return err
		}
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
