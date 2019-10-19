package storage

import (
	"context"
	"github.com/dgraph-io/badger"
	"github.com/gobwas/glob"
	"github.com/joway/pikv/types"
	"io"
	"regexp"
	"runtime"
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
	var output []byte = nil
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
	if err == badger.ErrKeyNotFound {
		return nil, types.ErrKeyNotFound
	}

	return output, err
}

func (storage *BadgerStorage) Set(key, val []byte, ttl uint64) error {
	return storage.db.Update(func(txn *badger.Txn) error {
		if ttl == 0 {
			return txn.Set(key, val)
		} else {
			e := badger.
				NewEntry(key, val).
				WithTTL(time.Millisecond * time.Duration(ttl))
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

func (storage *BadgerStorage) Scan(scanOpts ScanOptions) ([]KVPair, error) {
	var output []KVPair
	err := storage.db.View(func(txn *badger.Txn) error {
		//TODO: tuning prefetchSize
		opts := badger.IteratorOptions{
			PrefetchValues: scanOpts.IncludeValue,
			PrefetchSize:   runtime.GOMAXPROCS(0),
			Reverse:        false,
			AllVersions:    false,
		}
		it := txn.NewIterator(opts)
		defer it.Close()

		//check is prefix search
		var prefix []byte
		rePrefix := regexp.MustCompile(`^[\w]+\*$`)
		if rePrefix.MatchString(scanOpts.Pattern) {
			prefix = []byte(scanOpts.Pattern[:len(scanOpts.Pattern)-1])
		}
		globKey, err := glob.Compile(scanOpts.Pattern)
		if err != nil {
			return err
		}

		start := func(it *badger.Iterator) {
			if prefix == nil {
				it.Rewind()
			} else {
				it.Seek(prefix)
			}
		}
		valid := func(it *badger.Iterator) bool {
			//hit prefix optimization
			if prefix != nil {
				return it.ValidForPrefix(prefix)
			}

			return it.Valid()
		}
		for start(it); valid(it); it.Next() {
			if scanOpts.Limit > 0 && len(output) >= scanOpts.Limit {
				return nil
			}
			if scanOpts.Pattern != "" && !globKey.Match(string(it.Item().Key())) {
				continue
			}

			var pair = KVPair{}
			item := it.Item()
			pair.SetKey(item.KeyCopy(nil))
			if scanOpts.IncludeValue {
				v, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				pair.SetVal(v)
			}
			output = append(output, pair)
		}
		return nil
	})
	return output, err
}

func (storage *BadgerStorage) Snapshot(ctx context.Context, writer io.Writer) error {
	_, err := storage.db.Backup(writer, 0)
	if err != nil {
		return err
	}
	return nil
}

func (storage *BadgerStorage) LoadSnapshot(ctx context.Context, reader io.Reader) error {
	//TODO: custom maxPendingWrites
	return storage.db.Load(reader, 256)
}

func (storage *BadgerStorage) TTL(key []byte) (uint64, error) {
	var ttl uint64 = 0
	err := storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		ttl = item.ExpiresAt()
		if ttl > 0 {
			now := uint64(time.Now().Unix())
			//to milliseconds
			ttl = (ttl - now) * 1000
			if ttl <= 0 {
				//if key existed but ttl <= 0, return key not found error
				return types.ErrKeyNotFound
			}
		}
		//if not set ttl on key, return ttl = 0
		return nil
	})
	if err == badger.ErrKeyNotFound {
		return 0, types.ErrKeyNotFound
	}
	return ttl, err
}
