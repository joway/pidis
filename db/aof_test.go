package db

import (
	"bytes"
	"context"
	"fmt"
	"github.com/joway/pikv/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path"
	"testing"
	"time"
)

type AOFTestSuite struct {
	suite.Suite

	dir string
}

func TestAOF(t *testing.T) {
	suite.Run(t, new(AOFTestSuite))
}

func (suite *AOFTestSuite) SetupTest() {
	suite.dir = "/tmp/pikv/aof"
	_ = os.RemoveAll(suite.dir)
	_ = os.MkdirAll(suite.dir, os.ModePerm)
}

func (suite *AOFTestSuite) TestDecode() {
	text := []byte("*4\r\n$12\r\n000000000000\r\n$3\r\nset\r\n$1\r\nk\r\n$1\r\nv\r\n" +
		"*3\r\n$12\r\n000000000000\r\n$3\r\nget\r\n$1\r\nk\r\n")
	uid, args, leftover, err := DecodeAOF(text)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "000000000000", string(uid))
	assert.Equal(suite.T(), "*3\r\n$12\r\n000000000000\r\n$3\r\nget\r\n$1\r\nk\r\n", string(leftover))
	assert.Equal(suite.T(), "set k v", string(bytes.Join(args, []byte(" "))))
}

func (suite *AOFTestSuite) TestEncode() {
	uid := NewUID().Bytes()
	args := [][]byte{[]byte("set"), []byte("k"), []byte("v")}
	encoded := EncodeAOF(uid, args)
	assert.Equal(suite.T(), fmt.Sprintf("*4\r\n$12\r\n%s\r\n$3\r\nset\r\n$1\r\nk\r\n$1\r\nv\r\n", uid), string(encoded))
}

func (suite *AOFTestSuite) TestSync() {
	bus, err := NewAOFBus(path.Join(suite.dir, "test.aof"), UIDSize)
	assert.NoError(suite.T(), err)
	var offset []byte
	for i := 0; i < 100; i++ {
		if i == 50 {
			offset = NewUID().Bytes()
		}
		args := util.CommandToArgs(fmt.Sprintf("set k%d xxx", i))
		err := bus.Append(args)
		assert.NoError(suite.T(), err)
	}

	suite.NoError(bus.Flush())

	stream := util.NewStreamBus()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			if err := bus.Sync(ctx, stream, offset); err != nil {
				assert.NoError(suite.T(), err)
			}
		}
	}()

	total := 49
	for {
		select {
		case <-ctx.Done():
			assert.Equal(suite.T(), 99, total)
			return
		case content := <-stream.Read():
			for {
				ts, args, leftover, err := DecodeAOF(content)
				assert.NoError(suite.T(), err)
				if ts == nil && args == nil {
					break
				}
				total += 1
				assert.True(suite.T(), bytes.Compare(offset, ts) <= 0)
				assert.Equal(suite.T(), string(args[1]), fmt.Sprintf("k%d", total))
				content = leftover
			}
		}
	}
}
