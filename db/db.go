package db

import (
	"bufio"
	"fmt"
	"github.com/joway/pikv/storage"
	"github.com/pkg/errors"
	"os"
)

type Options struct {
	DBDir string
}

type Database struct {
	storage storage.Storage

	aofFile   *os.File
	aofWriter *AOFWriter
}

func NewDatabase(options Options) (*Database, error) {
	dataDir := fmt.Sprintf("%s/data", options.DBDir)
	aofFilePath := fmt.Sprintf("%s/pikv.aofWriter", options.DBDir)
	storageOpts := storage.Options{
		Storage: storage.TypeBadger,
		Dir:     dataDir,
	}
	store, err := storage.NewStorage(storageOpts)
	if err != nil {
		return nil, err
	}

	//create aofWriter stream
	aofFile, err := os.OpenFile(aofFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	aofBuffer := bufio.NewWriter(aofFile)
	aofWriter := NewAOFWriter(aofBuffer)

	return &Database{
		storage:   store,
		aofFile:   aofFile,
		aofWriter: aofWriter,
	}, nil
}

func (db *Database) Record(cmd [][]byte) error {
	return db.aofWriter.Append(cmd)
}

func (db *Database) Close() error {
	var errs []error
	if err := db.aofWriter.Flush(); err != nil {
		errs = append(errs, err)
	}
	if err := db.aofFile.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := db.storage.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Errorf("%v", errs)
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
