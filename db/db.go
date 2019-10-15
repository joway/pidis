package db

import (
	"context"
	"fmt"
	"github.com/joway/pikv/command"
	"github.com/joway/pikv/common"
	"github.com/joway/pikv/parser"
	"github.com/joway/pikv/storage"
	"github.com/joway/pikv/types"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"io"
	"os"
	"strings"
)

type Options struct {
	DBDir string
}

type Database struct {
	storage storage.Storage

	following *Node // slave of

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

func (db *Database) SlaveOf(host, port string) {
	db.following = &Node{
		host: host,
		port: port,
	}
}

func (db *Database) PullOplog(context context.Context, queue chan []byte, offset []byte) error {
	return db.aofBuf.Sync(context, queue, offset)
}

func (db *Database) Snapshot(writer io.Writer) ([]byte, error) {
	offset := xid.New().Bytes()
	if err := db.storage.Snapshot(writer); err != nil {
		return nil, err
	}
	return offset, nil
}

func (db *Database) LoadSnapshot(reader io.Reader) error {
	return db.storage.LoadSnapshot(reader)
}
