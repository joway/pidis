package db

import (
	"context"
	"fmt"
	"github.com/joway/loki"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/executor"
	"github.com/joway/pikv/rpc/proto"
	"github.com/joway/pikv/storage"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"io"
	"os"
	"path"
	"strings"
)

var logger = loki.New("db")

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

func (db *Database) Get(key []byte) ([]byte, error) {
	return db.storage.Get(key)
}
func (db *Database) Set(key, val []byte, ttl int64) error {
	return db.storage.Set(key, val, ttl)
}
func (db *Database) Del(keys [][]byte) error {
	return db.storage.Del(keys)
}

func (db *Database) IsWritable() bool {
	return db.following == nil
}

func (db *Database) Record(cmd [][]byte) error {
	return db.aofBus.Append(cmd)
}

func (db *Database) Close() error {
	var errs []error
	if err := db.followingConn.Close(); err != nil {
		errs = append(errs, err)
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

func (db *Database) exec(args [][]byte, isInternal bool) (output []byte, err error) {
	if len(args) == 0 {
		return nil, common.ErrInvalidNumberOfArgs
	}
	cmd := strings.ToUpper(string(args[0]))
	exec := executor.New(cmd)
	if exec.IsWrite() {
		if !isInternal && !db.IsWritable() {
			return nil, common.ErrNodeReadOnly
		}
		if err := db.Record(args); err != nil {
			return nil, err
		}
	}

	output, err = exec.Exec(db, args)
	return output, err
}

func (db *Database) Exec(args [][]byte) (output []byte, err error) {
	return db.exec(args, false)
}

func (db *Database) IExec(args [][]byte) (output []byte, err error) {
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
		case sig := <-db.sigFollowing:
			if sig {
				go func() {
					defer func() {
						db.sigFollowing <- false
					}()
					logger.Info("following node %s", db.following)

					fCtx, fCancel = context.WithCancel(ctx)
					if err := db.Following(fCtx); err != nil {
						logger.Fatal("%v", err)
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

func (db *Database) Following(ctx context.Context) error {
	if db.following == nil {
		return common.ErrNodeIsMaster
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
		if err := db.LoadSnapshot(ctx, snapFile); err != nil {
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
			_, args, _, err := Decode(line)
			if err != nil {
				return errors.Wrap(err, "parse oplog failed")
			}
			_, err = db.IExec(args)
			if err != nil {
				return errors.Wrap(err, "replay oplog failed")
			}
		}
	}
}

func (db *Database) Sync(ctx context.Context, writer io.Writer, offset []byte) error {
	return db.aofBus.Sync(ctx, writer, offset)
}

func (db *Database) Snapshot(ctx context.Context, writer io.Writer) error {
	if err := db.storage.Snapshot(ctx, writer); err != nil {
		return err
	}
	return nil
}

func (db *Database) LoadSnapshot(ctx context.Context, reader io.Reader) error {
	return db.storage.LoadSnapshot(ctx, reader)
}
