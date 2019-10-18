package db

import (
	"context"
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv/executor"
	"github.com/joway/pikv/proto"
	"github.com/joway/pikv/storage"
	"github.com/joway/pikv/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

var logger = loki.New("pikv:db")

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

	aofBus *AOFBus
}

func New(options Options) (*Database, error) {
	dataDir := path.Join(options.DBDir, "data")
	aofFilePath := path.Join(options.DBDir, "pikv.aof")
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

	//create aofBus stream
	aofBuf, err := NewAOFBus(aofFilePath, UIDSize)
	if err != nil {
		return nil, err
	}

	database := &Database{
		dir:     options.DBDir,
		storage: store,

		sigFollowing: make(chan bool),

		aofBus: aofBuf,
	}

	return database, nil
}

func (db *Database) Run() {
	go func() {
		if err := db.Daemon(); err != nil {
			logger.Fatal("Daemon dead: %v", err)
		}
	}()
}

func (db *Database) IsWritable() bool {
	return db.following == nil
}

func (db *Database) Record(cmd [][]byte) error {
	return db.aofBus.Append(cmd)
}

func (db *Database) Close() error {
	//TODO: improve error handle
	var errs []error
	if db.followingConn != nil {
		if err := db.followingConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := db.aofBus.Flush(); err != nil {
		errs = append(errs, err)
	}
	if err := db.aofBus.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := db.storage.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Errorf("%v", errs)
}

func (db *Database) exec(args [][]byte, isInternal bool) (result *executor.Result, err error) {
	if len(args) == 0 {
		return nil, types.ErrInvalidNumberOfArgs
	}
	cmd := strings.ToUpper(string(args[0]))
	exec := executor.New(cmd)
	if exec.IsWrite() {
		if !isInternal && !db.IsWritable() {
			return nil, types.ErrNodeReadOnly
		}
		if err := db.Record(args); err != nil {
			return nil, errors.Wrap(err, "record cmd failed")
		}
	}

	return exec.Exec(db.storage, args)
}

func (db *Database) Exec(args [][]byte) (result *executor.Result, err error) {
	return db.exec(args, false)
}

func (db *Database) IExec(args [][]byte) (result *executor.Result, err error) {
	return db.exec(args, true)
}

func (db *Database) Daemon() error {
	ctx := context.Background()
	var (
		fCtx    context.Context
		fCancel context.CancelFunc
	)

	for {
		select {
		//flush aof
		case <-time.After(time.Millisecond * 500):
			if err := db.aofBus.Flush(); err != nil {
				logger.Error("failed to flush aof file: %v", err)
			}
		case sig := <-db.sigFollowing:
			if sig {
				go func() {
					defer func() {
						db.sigFollowing <- false
					}()
					logger.Info("following node %s", db.following)

					fCtx, fCancel = context.WithCancel(ctx)
					if err := db.Following(fCtx); err != nil {
						logger.Error("failed to following: %v", err)
					}
				}()
			} else {
				logger.Info("removed following node %s", db.following)
				fCancel()
			}
		}
	}
}

func (db *Database) SlaveOf(host, port string) error {
	address := fmt.Sprintf("%s:%s", host, port)
	var err error
	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	db.followingConn, err = grpc.DialContext(
		ctx,
		address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return types.ErrNodeConnectFailed
	}
	db.following = &Node{
		host: host,
		port: port,
	}
	db.sigFollowing <- true
	return nil
}

func (db *Database) Following(ctx context.Context) error {
	if db.following == nil {
		return types.ErrNodeIsMaster
	}

	client := proto.NewPiKVClient(db.followingConn)
	offsetId := NewUID()
	//fetch snapshot
	snapPath := path.Join(db.dir, fmt.Sprintf("%s.snapshot", offsetId.Timestamp()))
	snapStream, err := client.Snapshot(ctx, &proto.SnapshotReq{})
	if err != nil {
		return errors.Wrap(err, "create snapshot stream failed")
	}
	snapFile, err := os.OpenFile(snapPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "cannot open snapshot file")
	}
	//save snapshot
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := snapStream.Recv()
			if err == io.EOF {
				goto restore
			}
			if err != nil {
				return errors.Wrap(err, "recv snapshot failed")
			}

			content := resp.GetPayload()
			if _, err := snapFile.Write(content); err != nil {
				return errors.Wrap(err, "append snapshot file failed")
			}
		}
	}
restore:
	{
		//restore snapshot
		_, _ = snapFile.Seek(0, io.SeekStart)
		if err := db.storage.LoadSnapshot(ctx, snapFile); err != nil {
			return errors.Wrap(err, "load snapshot failed")
		}
	}

	//fetch and replay oplog
	oplogStream, err := client.Oplog(ctx, &proto.OplogReq{
		Offset: offsetId.Bytes(),
	})
	if err != nil {
		return errors.Wrap(err, "fetch oplog failed")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := oplogStream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return errors.Wrap(err, "recv oplog failed")
			}

			line := resp.GetPayload()
			//TODO: concurrent
			//replay oplog
			for {
				_, args, leftover, err := DecodeAOF(line)
				if err != nil {
					return errors.Wrap(err, "parse oplog failed")
				}
				_, err = db.IExec(args)
				if err != nil {
					return errors.Wrap(err, "replay oplog failed")
				}
				if len(leftover) == 0 {
					break
				}
				line = leftover
			}
		}
	}
}

func (db *Database) Sync(ctx context.Context, writer io.Writer, offset []byte) error {
	return db.aofBus.Sync(ctx, writer, offset)
}
