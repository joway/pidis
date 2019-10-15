package db

import (
	"bufio"
	"github.com/joway/pikv/types"
	"github.com/joway/pikv/util"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func setup() {
	_ = os.RemoveAll("/tmp/pikv")
}

func TestDatabase_Snapshot(t *testing.T) {
	setup()

	db, err := NewDatabase(Options{
		DBDir: "/tmp/pikv",
	})
	assert.NoError(t, err)

	_, _, err = db.Exec(types.Context{
		Out:  nil,
		Args: util.CommandToArgs("set a x"),
	})
	assert.NoError(t, err)

	f, err := os.OpenFile("/tmp/pikv/pikv.snap", os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer f.Close()
	writer := bufio.NewWriter(f)
	offset, err := db.Snapshot(writer)
	assert.NoError(t, err)
	err = writer.Flush()
	assert.NoError(t, err)
	assert.NotEmpty(t, offset)

	newDb, err := NewDatabase(Options{
		DBDir: "/tmp/pikv/new",
	})
	assert.NoError(t, err)
	_, err = f.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	reader := bufio.NewReader(f)
	err = newDb.LoadSnapshot(reader)
	assert.NoError(t, err)
	output, _, err := newDb.Exec(types.Context{
		Out:  nil,
		Args: util.CommandToArgs("get a"),
	})
	assert.Equal(t, output[4], byte('x'))
}
