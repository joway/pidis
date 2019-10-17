package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var tmpStorageDir = "/tmp/pikv/storage"

func setup() {
	_ = os.RemoveAll(tmpStorageDir)
}

func TestBadgerStorage(t *testing.T) {
	setup()

	badgerStorage, err := NewBadgerStorage(Options{Dir: tmpStorageDir})
	assert.NoError(t, err)
	testStorage(t, badgerStorage)
}

func TestMemoryStorage(t *testing.T) {
	setup()

	memoryStorage, err := NewMemoryStorage(Options{})
	assert.NoError(t, err)
	testStorage(t, memoryStorage)
}

func TestBadgerStorageWithTTL(t *testing.T) {
	setup()

	badgerStorage, err := NewBadgerStorage(Options{Dir: tmpStorageDir})
	assert.NoError(t, err)
	testStorageWithTTL(t, badgerStorage)
}

func TestMemoryStorageWithTTL(t *testing.T) {
	setup()

	memoryStorage, err := NewMemoryStorage(Options{})
	assert.NoError(t, err)
	testStorageWithTTL(t, memoryStorage)
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
		assert.Equal(t, string(v), fmt.Sprintf("%d", i))
	}

	v, err := storage.Get([]byte("k100"))
	assert.NoError(t, err)
	assert.Nil(t, v)

	err = storage.Del([][]byte{[]byte("k99")})
	assert.NoError(t, err)
	v, err = storage.Get([]byte("k99"))
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func testStorageWithTTL(t *testing.T, storage Storage) {
	setup()

	err := storage.Set([]byte("k"), []byte("xxx"), 1000)
	assert.NoError(t, err)
	v, err := storage.Get([]byte("k"))
	assert.Equal(t, string(v), "xxx")
	time.Sleep(time.Second)
	v, err = storage.Get([]byte("k"))
	assert.Nil(t, v)
}
