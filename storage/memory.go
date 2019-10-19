package storage

import (
	"github.com/gobwas/glob"
	"github.com/joway/pikv/types"
	"github.com/tidwall/buntdb"
	"time"
)

type MemoryStorage struct {
	Storage

	db *buntdb.DB
}

func NewMemoryStorage(options Options) (Storage, error) {
	db, err := buntdb.Open(":memory:")
	return &MemoryStorage{db: db}, err
}

func (storage *MemoryStorage) Close() error {
	return storage.db.Close()
}

func (storage *MemoryStorage) Get(key []byte) ([]byte, error) {
	var output []byte
	err := storage.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(string(key))
		if err == buntdb.ErrNotFound {
			return types.ErrKeyNotFound
		}
		output = []byte(val)
		return err
	})
	return output, err
}

func (storage *MemoryStorage) Set(key, val []byte, ttl uint64) error {
	var opts *buntdb.SetOptions
	if ttl == 0 {
		opts = nil
	} else {
		opts = &buntdb.SetOptions{
			Expires: true,
			TTL:     time.Millisecond * time.Duration(ttl),
		}
	}
	err := storage.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(string(key), string(val), opts)
		return err
	})
	return err
}

func (storage *MemoryStorage) Del(keys [][]byte) error {
	for _, key := range keys {
		err := storage.db.Update(func(tx *buntdb.Tx) error {
			_, err := tx.Delete(string(key))
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (storage *MemoryStorage) Scan(scanOpts ScanOptions) ([]KVPair, error) {
	var output []KVPair
	reGlob, err := glob.Compile(scanOpts.Pattern)
	if err != nil {
		return nil, err
	}
	err = storage.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			if scanOpts.Limit > 0 && len(output) >= scanOpts.Limit {
				return false
			}
			if scanOpts.Pattern != "" {
				//skip
				if !reGlob.Match(key) {
					return true
				}
			}

			pair := KVPair{}
			pair.SetKey([]byte(key))
			if scanOpts.IncludeValue {
				pair.SetVal([]byte(value))
			}
			output = append(output, pair)
			return true
		})
		return err
	})

	return output, err
}

func (storage *MemoryStorage) TTL(key []byte) (uint64, error) {
	var ttl uint64
	err := storage.db.View(func(tx *buntdb.Tx) error {
		exp, err := tx.TTL(string(key))
		if err == buntdb.ErrNotFound {
			return types.ErrKeyNotFound
		}
		if err != nil {
			return err
		}
		if exp == 0 {
			return types.ErrKeyNotFound
		} else if exp < 0 {
			ttl = 0
		} else if exp > 0 {
			ttl = uint64(exp.Milliseconds())
		}
		return err
	})
	return ttl, err
}
