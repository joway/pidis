package storage

import (
	"fmt"
	"github.com/joway/pikv/types"
	"github.com/stretchr/testify/assert"
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
	testStorage(suite.T(), badgerStorage)
}

func (suite *StorageTestSuite) TestMemoryStorage() {
	memoryStorage, err := NewMemoryStorage(Options{})
	suite.NoError(err)
	testStorage(suite.T(), memoryStorage)
}

func (suite *StorageTestSuite) TestBadgerStorageWithTTL() {
	badgerStorage, err := NewBadgerStorage(Options{Dir: suite.dir})
	suite.NoError(err)
	testStorageWithTTL(suite.T(), badgerStorage)
}

func (suite *StorageTestSuite) TestMemoryStorageWithTTL() {
	memoryStorage, err := NewMemoryStorage(Options{})
	suite.NoError(err)
	testStorageWithTTL(suite.T(), memoryStorage)
}

func (suite *StorageTestSuite) TestBadgerStorage_Scan() {
	badgerStorage, err := NewBadgerStorage(Options{Dir: suite.dir})
	suite.NoError(err)
	testStorageScan(suite.T(), badgerStorage)
}

func (suite *StorageTestSuite) TestMemoryStorage_Scan() {
	memoryStorage, err := NewMemoryStorage(Options{})
	suite.NoError(err)
	testStorageScan(suite.T(), memoryStorage)
}

func testStorage(t *testing.T, storage Storage) {
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("%d", i)
		err := storage.Set([]byte(k), []byte(v), 0)
		assert.NoError(t, err)
	}

	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v, err := storage.Get([]byte(k))
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d", i), string(v))
	}

	v, err := storage.Get([]byte("k100"))
	assert.Equal(t, types.ErrKeyNotFound, err)
	assert.Nil(t, v)

	err = storage.Del([][]byte{[]byte("k99")})
	assert.NoError(t, err)
	v, err = storage.Get([]byte("k99"))
	assert.Equal(t, types.ErrKeyNotFound, err)
	assert.Nil(t, v)
}

func testStorageWithTTL(t *testing.T, storage Storage) {
	err := storage.Set([]byte("k"), []byte("xxx"), 2000)
	assert.NoError(t, err)
	ttl, err := storage.TTL([]byte("k"))
	assert.NoError(t, err)
	assert.LessOrEqual(t, ttl, uint64(2000))
	v, err := storage.Get([]byte("k"))
	assert.Equal(t, "xxx", string(v))

	time.Sleep(time.Millisecond * 1000)
	ttl, err = storage.TTL([]byte("k"))
	assert.NoError(t, err)
	assert.LessOrEqual(t, ttl, uint64(1000))

	time.Sleep(time.Millisecond * 1000)
	v, err = storage.Get([]byte("k"))
	assert.Nil(t, v)
}

func testStorageScan(t *testing.T, storage Storage) {
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("%d", i)
		err := storage.Set([]byte(k), []byte(v), 0)
		assert.NoError(t, err)
	}

	pairs, err := storage.Scan(ScanOptions{Limit: 1000, IncludeValue: true})
	assert.NoError(t, err)
	assert.Equal(t, 100, len(pairs))
	for _, pair := range pairs {
		assert.NotNil(t, pair.Val)
	}

	pairs, err = storage.Scan(ScanOptions{Limit: 10, IncludeValue: false})
	assert.NoError(t, err)
	assert.Equal(t, 10, len(pairs))
	for _, pair := range pairs {
		assert.Nil(t, pair.Val)
	}

	pairs, err = storage.Scan(ScanOptions{Limit: -1, IncludeValue: false, Pattern: "k1"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pairs))

	pairs, err = storage.Scan(ScanOptions{Limit: -1, IncludeValue: false, Pattern: "k1*"})
	assert.NoError(t, err)
	assert.Equal(t, 11, len(pairs))
}
