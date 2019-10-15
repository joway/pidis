package db

import (
	"bufio"
	"bytes"
	"context"
	"github.com/rs/xid"
	"io"
	"os"
)

const (
	OffsetSize = 12
)

func AOFEncode(uuid []byte, args [][]byte) []byte {
	fullCmd := bytes.Join(args, []byte(" "))
	return append(append(uuid, fullCmd...), '\n')
}

func AOFDecode(line []byte) (offset []byte, args [][]byte) {
	offset = line[:OffsetSize]
	args = bytes.Split(line[OffsetSize:], []byte(" "))
	return offset, args
}

type AOFBus struct {
	filePath string

	appendFile   *os.File
	appendBuffer *bufio.Writer
}

func NewAOFBus(filePath string) (*AOFBus, error) {
	appendFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	appendBuffer := bufio.NewWriter(appendFile)

	return &AOFBus{
		filePath: filePath,

		appendFile:   appendFile,
		appendBuffer: appendBuffer,
	}, nil
}

func (w *AOFBus) Append(cmd [][]byte) error {
	guid := xid.New()
	line := AOFEncode(guid.Bytes(), cmd)
	if _, err := w.appendBuffer.Write(line); err != nil {
		return err
	}
	return w.Flush()
}

func (w *AOFBus) Flush() error {
	return w.appendBuffer.Flush()
}

func (w *AOFBus) Close() error {
	return w.appendFile.Close()
}

func (w *AOFBus) Sync(context context.Context, writer io.Writer, offset []byte) error {
	aofFile, err := os.OpenFile(w.filePath, os.O_RDONLY, 0600)
	defer aofFile.Close()
	if err != nil {
		return err
	}
	rd := bufio.NewReader(aofFile)
	for {
		select {
		case <-context.Done():
			return nil
		default:
			//TODO: care about isPrefix == true
			line, _, err := rd.ReadLine()
			if err != nil {
				return err
			}
			timestamp := line[:OffsetSize]
			if offset != nil && bytes.Compare(timestamp, offset) < 0 {
				continue
			}

			if _, err := writer.Write(line); err != nil {
				return err
			}
		}
	}
}
