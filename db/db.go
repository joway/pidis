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
	"github.com/rs/xid"
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

	aofBuf *AOFBus
}

func New(options Options) (*Database, error) {
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

		sigFollowing: make(chan bool),

		aofBuf: aofBuf,
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
	var followingContext context.Context
	for {
		select {
		case sig := <-db.sigFollowing:
			if sig {
				go func() {
					defer func() {
						db.sigFollowing <- false
					}()
					logger.Info("following node %s", db.following)
					followingContext = context.TODO()
					if err := db.Following(followingContext); err != nil {
						logger.Fatal("%v", err)
					}
				}()
			} else {
				logger.Info("removed following node %s", db.following)
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
		return common.ErrNodeIsMaster
	}

	client := proto.NewPiKVClient(db.followingConn)
	offsetId := xid.New()
	offset := xid.New().Bytes()
	//fetch snapshot
	snapPath := path.Join(db.dir, fmt.Sprintf("%s.snapshot", string(offsetId.Time().Unix())))
	snapStream, err := client.Snapshot(context, &proto.SnapshotReq{})
	if err != nil {
		return errors.Wrap(err, "create snapshot stream failed")
	}
	snapFile, err := os.OpenFile(snapPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "cannot open snapshot file")
	}
	//save snapshot
	for {
		resp, err := snapStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "recv snapshot failed")
		}

		content := resp.GetPayload()
		if _, err := snapFile.Write(content); err != nil {
			return errors.Wrap(err, "append snapshot file failed")
		}
	}
	//restore snapshot
	_, _ = snapFile.Seek(0, io.SeekStart)
	if err := db.LoadSnapshot(snapFile); err != nil {
		return errors.Wrap(err, "load snapshot failed")
	}
	//fetch and replay oplog
	oplogStream, err := client.Oplog(context, &proto.OplogReq{
		Offset: offset,
	})
	if err != nil {
		return errors.Wrap(err, "fetch oplog failed")
	}
	for {
		resp, err := oplogStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "recv oplog failed")
		}

		line := resp.GetPayload()
		fmt.Println("line", line)
		//replay oplog
		//TODO: concurrent
		_, args := AOFDecode(line)
		_, err = db.IExec(args)
		if err != nil {
			return errors.Wrap(err, "replay oplog failed")
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
