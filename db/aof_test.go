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

func TestDecode(t *testing.T) {
	text := []byte("*4\r\n$12\r\n000000000000\r\n$3\r\nset\r\n$1\r\nk\r\n$1\r\nv\r\n" +
		"*3\r\n$12\r\n000000000000\r\n$3\r\nget\r\n$1\r\nk\r\n")
	uid, args, leftover, err := DecodeAOF(text)
	assert.NoError(t, err)
	assert.Equal(t, "000000000000", string(uid))
	assert.Equal(t, "*3\r\n$12\r\n000000000000\r\n$3\r\nget\r\n$1\r\nk\r\n", string(leftover))
	assert.Equal(t, "set k v", string(bytes.Join(args, []byte(" "))))
}

func TestEncode(t *testing.T) {
	uid := NewUID().Bytes()
	args := [][]byte{[]byte("set"), []byte("k"), []byte("v")}
	encoded := EncodeAOF(uid, args)
	assert.Equal(t, fmt.Sprintf("*4\r\n$12\r\n%s\r\n$3\r\nset\r\n$1\r\nk\r\n$1\r\nv\r\n", uid), string(encoded))
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
		select {
		case <-ctx.Done():
			return
		default:
			if err := bus.Sync(ctx, stream, offset); err != nil {
				assert.NoError(t, err)
			}
		}
	}()

	total := 49
	for {
		select {
		case <-ctx.Done():
			assert.Equal(t, 99, total)
			return
		case content := <-stream.Read():
			for {
				ts, args, leftover, err := DecodeAOF(content)
				assert.NoError(t, err)
				if ts == nil && args == nil {
					break
				}
				total += 1
				assert.True(t, bytes.Compare(offset, ts) <= 0)
				assert.Equal(t, string(args[1]), fmt.Sprintf("k%d", total))
				content = leftover
			}
		}
	}
}
