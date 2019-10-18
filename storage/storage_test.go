package storage

import (
	"fmt"
	"github.com/joway/pikv/common"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var tmpStorageDir = "/tmp/pikv/storage"

func setup() {
	_ = os.RemoveAll(tmpStorageDir)
}

//
//func TestBadgerStorage(t *testing.T) {
//	setup()
//
//	badgerStorage, err := NewBadgerStorage(Options{Dir: tmpStorageDir})
//	assert.NoError(t, err)
//	testStorage(t, badgerStorage)
//}
//
//func TestMemoryStorage(t *testing.T) {
//	setup()
//
//	memoryStorage, err := NewMemoryStorage(Options{})
//	assert.NoError(t, err)
//	testStorage(t, memoryStorage)
//}
//
//func TestBadgerStorageWithTTL(t *testing.T) {
//	setup()
//
//	badgerStorage, err := NewBadgerStorage(Options{Dir: tmpStorageDir})
//	assert.NoError(t, err)
//	testStorageWithTTL(t, badgerStorage)
//}
//
//func TestMemoryStorageWithTTL(t *testing.T) {
//	setup()
//
//	memoryStorage, err := NewMemoryStorage(Options{})
//	assert.NoError(t, err)
//	testStorageWithTTL(t, memoryStorage)
//}

func TestBadgerStorage_Scan(t *testing.T) {
	setup()

	badgerStorage, err := NewBadgerStorage(Options{Dir: tmpStorageDir})
	assert.NoError(t, err)
	testStorageScan(t, badgerStorage)
}

func TestMemoryStorage_Scan(t *testing.T) {
	setup()

	memoryStorage, err := NewMemoryStorage(Options{})
	assert.NoError(t, err)
	testStorageScan(t, memoryStorage)
}

func testStorage(t *testing.T, storage Storage) {
	setup()

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
	assert.Equal(t, common.ErrKeyNotFound, err)
	assert.Nil(t, v)

	err = storage.Del([][]byte{[]byte("k99")})
	assert.NoError(t, err)
	v, err = storage.Get([]byte("k99"))
	assert.Equal(t, common.ErrKeyNotFound, err)
	assert.Nil(t, v)
}

func testStorageWithTTL(t *testing.T, storage Storage) {
	setup()

	err := storage.Set([]byte("k"), []byte("xxx"), 1000)
	assert.NoError(t, err)
	v, err := storage.Get([]byte("k"))
	assert.Equal(t, "xxx", string(v))
	time.Sleep(time.Second)
	v, err = storage.Get([]byte("k"))
	assert.Nil(t, v)
}

func testStorageScan(t *testing.T, storage Storage) {
	setup()

	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("%d", i)
		err := storage.Set([]byte(k), []byte(v), 0)
		assert.NoError(t, err)
	}

	pairs, err := storage.Scan(common.ScanOptions{Limit: 1000, IncludeValue: true})
	assert.NoError(t, err)
	assert.Equal(t, 100, len(pairs))
	for _, pair := range pairs {
		assert.NotNil(t, pair.Val)
	}

	pairs, err = storage.Scan(common.ScanOptions{Limit: 10, IncludeValue: false})
	assert.NoError(t, err)
	assert.Equal(t, 10, len(pairs))
	for _, pair := range pairs {
		assert.Nil(t, pair.Val)
	}

	pairs, err = storage.Scan(common.ScanOptions{Limit: -1, IncludeValue: false, Pattern: "k1"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pairs))

	pairs, err = storage.Scan(common.ScanOptions{Limit: -1, IncludeValue: false, Pattern: "k1*"})
	assert.NoError(t, err)
	assert.Equal(t, 11, len(pairs))
}
