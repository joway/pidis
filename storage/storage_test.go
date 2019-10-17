package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func setup() {
	_ = os.RemoveAll("/tmp/pikv")
}

var badgerStorage, _ = NewBadgerStorage(Options{Dir: "/tmp/pikv"})
var memoryStorage, _ = NewMemoryStorage(Options{})
var storageList = []Storage{badgerStorage, memoryStorage}

func TestStorage(t *testing.T) {
	setup()

	for _, storage := range storageList {
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
	}
}

func TestStorage_TTL(t *testing.T) {
	setup()

	for _, storage := range storageList {
		err := storage.Set([]byte("k"), []byte("xxx"), 1000)
		assert.NoError(t, err)
		v, err := storage.Get([]byte("k"))
		assert.Equal(t, string(v), "xxx")
		time.Sleep(time.Second)
		v, err = storage.Get([]byte("k"))
		assert.Nil(t, v)
	}
}
