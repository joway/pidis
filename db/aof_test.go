package db

import (
	"bytes"
	"context"
	"fmt"
	"github.com/joway/pikv/util"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"time"
)

var tmpAofDir = "/tmp/pikv/aof"

func setupAOFBus() {
	_ = os.RemoveAll(tmpAofDir)
	_ = os.MkdirAll(tmpAofDir, os.ModePerm)
}

func TestAOFBus(t *testing.T) {
	setupAOFBus()

	bus, err := NewAOFBus(path.Join(tmpAofDir, "test.aof"), UIDSize)
	assert.NoError(t, err)
	var offset []byte
	for i := 0; i < 100; i++ {
		if i == 50 {
			offset = NewUID().Bytes()
		}
		args := util.CommandToArgs(fmt.Sprintf("set k%d xxx", i))
		err := bus.Append(args)
		assert.NoError(t, err)
	}

	stream := util.NewStreamBus()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	go func() {
		if err := bus.Sync(ctx, stream, offset); err != nil {
			assert.NoError(t, err)
		}
	}()

	total := 49
	for {
		select {
		case <-ctx.Done():
			assert.Equal(t, total, 99)
			return
		case line := <-stream.Read():
			total += 1
			ts, args := bus.DecodeLine(line)
			assert.True(t, bytes.Compare(offset, ts) <= 0)
			assert.Equal(t, string(args[1]), fmt.Sprintf("k%d", total))
		}
	}
}
