package storage

import (
	"fmt"
	"github.com/joway/pikv/types"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type StorageTestSuite struct {
	suite.Suite

	dir string
}

func TestStorage(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (suite *StorageTestSuite) SetupTest() {
	suite.dir = "/tmp/pikv/storage"
	_ = os.RemoveAll(suite.dir)
}

func (suite *StorageTestSuite) TestBadgerStorage() {
	badgerStorage, err := NewBadgerStorage(Options{Dir: suite.dir})
	suite.NoError(err)
	testStorage(suite, badgerStorage)
}

func (suite *StorageTestSuite) TestMemoryStorage() {
	memoryStorage, err := NewMemoryStorage(Options{})
	suite.NoError(err)
	testStorage(suite, memoryStorage)
}

func (suite *StorageTestSuite) TestBadgerStorageWithTTL() {
	badgerStorage, err := NewBadgerStorage(Options{Dir: suite.dir})
	suite.NoError(err)
	testStorageWithTTL(suite, badgerStorage)
}

func (suite *StorageTestSuite) TestMemoryStorageWithTTL() {
	memoryStorage, err := NewMemoryStorage(Options{})
	suite.NoError(err)
	testStorageWithTTL(suite, memoryStorage)
}

func (suite *StorageTestSuite) TestBadgerStorage_Scan() {
	badgerStorage, err := NewBadgerStorage(Options{Dir: suite.dir})
	suite.NoError(err)
	testStorageScan(suite, badgerStorage)
}

func (suite *StorageTestSuite) TestMemoryStorage_Scan() {
	memoryStorage, err := NewMemoryStorage(Options{})
	suite.NoError(err)
	testStorageScan(suite, memoryStorage)
}

func testStorage(suite *StorageTestSuite, storage Storage) {
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("%d", i)
		err := storage.Set([]byte(k), []byte(v), 0)
		suite.NoError(err)
	}

	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v, err := storage.Get([]byte(k))
		suite.NoError(err)
		suite.Equal(fmt.Sprintf("%d", i), string(v))
	}

	v, err := storage.Get([]byte("k100"))
	suite.Equal(types.ErrKeyNotFound, err)
	suite.Nil(v)

	err = storage.Del([][]byte{[]byte("k99")})
	suite.NoError(err)
	v, err = storage.Get([]byte("k99"))
	suite.Equal(types.ErrKeyNotFound, err)
	suite.Nil(v)
}

func testStorageWithTTL(suite *StorageTestSuite, storage Storage) {
	err := storage.Set([]byte("k"), []byte("xxx"), 2000)
	suite.NoError(err)
	ttl, err := storage.TTL([]byte("k"))
	suite.NoError(err)
	suite.LessOrEqual(ttl, uint64(2000))
	v, err := storage.Get([]byte("k"))
	suite.NoError(err)
	suite.Equal("xxx", string(v))

	time.Sleep(time.Millisecond * 1000)
	ttl, err = storage.TTL([]byte("k"))
	suite.NoError(err)
	suite.LessOrEqual(ttl, uint64(1000))

	time.Sleep(time.Millisecond * 1000)
	v, err = storage.Get([]byte("k"))
	suite.Equal(types.ErrKeyNotFound, err)
	suite.Nil(v)
}

func testStorageScan(suite *StorageTestSuite, storage Storage) {
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("%d", i)
		err := storage.Set([]byte(k), []byte(v), 0)
		suite.NoError(err)
	}

	pairs, err := storage.Scan(ScanOptions{Limit: 1000, IncludeValue: true})
	suite.NoError(err)
	suite.Equal(100, len(pairs))
	for _, pair := range pairs {
		suite.NotNil(pair.Val)
	}

	pairs, err = storage.Scan(ScanOptions{Limit: 10, IncludeValue: false})
	suite.NoError(err)
	suite.Equal(10, len(pairs))
	for _, pair := range pairs {
		suite.Nil(pair.Val)
	}

	pairs, err = storage.Scan(ScanOptions{Limit: -1, IncludeValue: false, Pattern: "k1"})
	suite.NoError(err)
	suite.Equal(1, len(pairs))

	pairs, err = storage.Scan(ScanOptions{Limit: -1, IncludeValue: false, Pattern: "k1*"})
	suite.NoError(err)
	suite.Equal(11, len(pairs))
}
