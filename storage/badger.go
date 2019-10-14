package storage

import (
	"github.com/dgraph-io/badger"
	"time"
)

type BadgerStorage struct {
	Storage

	db *badger.DB
}

func NewBadgerStorage(options Options) (Storage, error) {
	db, err := badger.Open(badger.DefaultOptions(options.Dir))
	if err != nil {
		return nil, err
	}

	return &BadgerStorage{db: db}, nil
}

func (storage *BadgerStorage) Close() error {
	return storage.db.Close()
}

func (storage *BadgerStorage) Get(key []byte) ([]byte, error) {
	var output []byte
	err := storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		output = val

		return nil
	})

	return output, err
}

func (storage *BadgerStorage) Set(key, val []byte, ttl int64) error {
	return storage.db.Update(func(txn *badger.Txn) error {
		if ttl < 1 {
			return txn.Set(key, val)
		} else {
			e := badger.NewEntry(key, val).WithTTL(time.Duration(ttl) * time.Millisecond)
			return txn.SetEntry(e)
		}
	})
}

func (storage *BadgerStorage) Del(keys [][]byte) error {
	return storage.db.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})
}
