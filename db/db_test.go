package db

import (
	"bufio"
	"context"
	"github.com/joway/pidis/util"
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
	suite.dir = "/tmp/pidis/db"
	_ = os.RemoveAll(suite.dir)
	_ = os.MkdirAll(suite.dir, os.ModePerm)
}

func (suite *DBTestSuite) TestSlaveOf() {
	leader, err := New(Options{
		DBDir: path.Join(suite.dir, "leader"),
	})
	suite.NoError(err)
	follower, err := New(Options{
		DBDir: path.Join(suite.dir, "follower"),
	})
	suite.NoError(err)
	leader.Run()
	follower.Run()

	leaderListen, err := net.Listen("tcp", ":10001")
	suite.NoError(err)
	go func() {
		server := NewRpcServer(leader)
		err := server.Serve(leaderListen)
		suite.NoError(err)
	}()

	_, err = leader.Exec(util.CommandToArgs("set k x"))
	suite.NoError(err)
	result, err := leader.Exec(util.CommandToArgs("get k"))
	suite.NoError(err)
	suite.Equal(result.Output()[4], byte('x'))

	err = follower.SlaveOf("0.0.0.0", "10001")
	suite.NoError(err)

	time.Sleep(time.Second)

	_, err = leader.Exec(util.CommandToArgs("set k1 xxx"))
	suite.NoError(err)
	_, err = leader.Exec(util.CommandToArgs("set k2 xxx"))
	suite.NoError(err)
	time.Sleep(time.Millisecond * 100)
	_, err = leader.Exec(util.CommandToArgs("set k3 xxx"))
	suite.NoError(err)

	time.Sleep(time.Millisecond * 1000)

	result, err = follower.Exec(util.CommandToArgs("get k"))
	suite.NoError(err)
	suite.Equal(byte('x'), result.Output()[4])
	result, err = follower.Exec(util.CommandToArgs("get k1"))
	suite.NoError(err)
	suite.Equal("xxx", string(result.Output()[4:7]))
	result, err = follower.Exec(util.CommandToArgs("get k2"))
	suite.NoError(err)
	suite.Equal("xxx", string(result.Output()[4:7]))
	result, err = follower.Exec(util.CommandToArgs("get k3"))
	suite.NoError(err)
	suite.Equal("xxx", string(result.Output()[4:7]))
}

func (suite *DBTestSuite) TestSnapshot() {
	ctx := context.Background()

	db, err := New(Options{
		DBDir: suite.dir,
	})
	suite.NoError(err)

	_, err = db.Exec(util.CommandToArgs("set a x"))
	suite.NoError(err)

	f, err := os.OpenFile(path.Join(suite.dir, "pidis.snap"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	suite.NoError(err)
	defer func() { suite.NoError(f.Close()) }()
	writer := bufio.NewWriter(f)
	err = db.storage.Snapshot(ctx, writer)
	suite.NoError(err)
	err = writer.Flush()
	suite.NoError(err)

	newDb, err := New(Options{
		DBDir: path.Join(suite.dir, "new"),
	})
	suite.NoError(err)
	_, err = f.Seek(0, io.SeekStart)
	suite.NoError(err)
	err = newDb.storage.LoadSnapshot(ctx, f)
	suite.NoError(err)
	result, err := newDb.Exec(util.CommandToArgs("get a"))
	suite.NoError(err)
	suite.Equal(result.Output()[4], byte('x'))
}
