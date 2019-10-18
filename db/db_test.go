package db

import (
	"bufio"
	"context"
	"github.com/joway/pikv/util"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"os"
	"path"
	"testing"
	"time"
)

var tmpDbDir = "/tmp/pikv/db"

func setup() {
	_ = os.RemoveAll(tmpDbDir)
	_ = os.MkdirAll(tmpDbDir, os.ModePerm)
}

func TestDatabase_SlaveOf(t *testing.T) {
	setup()

	leader, err := New(Options{
		DBDir: path.Join(tmpDbDir, "leader"),
	})
	assert.NoError(t, err)
	follower, err := New(Options{
		DBDir: path.Join(tmpDbDir, "follower"),
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

	time.Sleep(time.Second)

	result, err = follower.Exec(util.CommandToArgs("get k"))
	assert.NoError(t, err)
	assert.Equal(t, byte('x'), result.Output()[4])
	result, err = follower.Exec(util.CommandToArgs("get k1"))
	assert.NoError(t, err)
	assert.Equal(t, "xxx", string(result.Output()[4:7]))
	result, err = follower.Exec(util.CommandToArgs("get k2"))
	assert.NoError(t, err)
	//assert.Equal(t, "xxx", string(result.Output()[4:7]))
}

func TestDatabase_Snapshot(t *testing.T) {
	setup()
	ctx := context.Background()

	db, err := New(Options{
		DBDir: tmpDbDir,
	})
	assert.NoError(t, err)

	_, err = db.Exec(util.CommandToArgs("set a x"))
	assert.NoError(t, err)

	f, err := os.OpenFile(path.Join(tmpDbDir, "pikv.snap"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer f.Close()
	writer := bufio.NewWriter(f)
	err = db.storage.Snapshot(ctx, writer)
	assert.NoError(t, err)
	err = writer.Flush()
	assert.NoError(t, err)

	newDb, err := New(Options{
		DBDir: path.Join(tmpDbDir, "new"),
	})
	assert.NoError(t, err)
	_, err = f.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	err = newDb.storage.LoadSnapshot(ctx, f)
	assert.NoError(t, err)
	result, err := newDb.Exec(util.CommandToArgs("get a"))
	assert.Equal(t, result.Output()[4], byte('x'))
}
