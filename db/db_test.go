package db

import (
	"bufio"
	"context"
	"github.com/joway/pikv/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net"
	"os"
	"path"
	"testing"
	"time"
)

type DBTestSuite struct {
	suite.Suite

	dir string
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

func (suite *DBTestSuite) SetupTest() {
	suite.dir = "/tmp/pikv/db"
	_ = os.RemoveAll(suite.dir)
	_ = os.MkdirAll(suite.dir, os.ModePerm)
}

func (suite *DBTestSuite) TestSlaveOf() {
	t := suite.T()
	leader, err := New(Options{
		DBDir: path.Join(suite.dir, "leader"),
	})
	assert.NoError(t, err)
	follower, err := New(Options{
		DBDir: path.Join(suite.dir, "follower"),
	})
	assert.NoError(t, err)
	leader.Run()
	follower.Run()

	leaderListen, err := net.Listen("tcp", ":10001")
	assert.NoError(t, err)
	go func() {
		server := NewRpcServer(leader)
		err := server.Serve(leaderListen)
		assert.NoError(t, err)
	}()

	_, err = leader.Exec(util.CommandToArgs("set k x"))
	assert.NoError(t, err)
	result, err := leader.Exec(util.CommandToArgs("get k"))
	assert.NoError(t, err)
	assert.Equal(t, result.Output()[4], byte('x'))

	err = follower.SlaveOf("0.0.0.0", "10001")
	assert.NoError(t, err)

	time.Sleep(time.Second)

	_, err = leader.Exec(util.CommandToArgs("set k1 xxx"))
	assert.NoError(t, err)
	_, err = leader.Exec(util.CommandToArgs("set k2 xxx"))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	_, err = leader.Exec(util.CommandToArgs("set k3 xxx"))
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 1000)

	result, err = follower.Exec(util.CommandToArgs("get k"))
	assert.NoError(t, err)
	assert.Equal(t, byte('x'), result.Output()[4])
	result, err = follower.Exec(util.CommandToArgs("get k1"))
	assert.NoError(t, err)
	assert.Equal(t, "xxx", string(result.Output()[4:7]))
	result, err = follower.Exec(util.CommandToArgs("get k2"))
	assert.NoError(t, err)
	assert.Equal(t, "xxx", string(result.Output()[4:7]))
	result, err = follower.Exec(util.CommandToArgs("get k3"))
	assert.NoError(t, err)
	assert.Equal(t, "xxx", string(result.Output()[4:7]))
}

func (suite *DBTestSuite) TestSnapshot() {
	t := suite.T()
	ctx := context.Background()

	db, err := New(Options{
		DBDir: suite.dir,
	})
	assert.NoError(t, err)

	_, err = db.Exec(util.CommandToArgs("set a x"))
	assert.NoError(t, err)

	f, err := os.OpenFile(path.Join(suite.dir, "pikv.snap"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer f.Close()
	writer := bufio.NewWriter(f)
	err = db.storage.Snapshot(ctx, writer)
	assert.NoError(t, err)
	err = writer.Flush()
	assert.NoError(t, err)

	newDb, err := New(Options{
		DBDir: path.Join(suite.dir, "new"),
	})
	assert.NoError(t, err)
	_, err = f.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	err = newDb.storage.LoadSnapshot(ctx, f)
	assert.NoError(t, err)
	result, err := newDb.Exec(util.CommandToArgs("get a"))
	assert.Equal(t, result.Output()[4], byte('x'))
}
