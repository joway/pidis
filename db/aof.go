package db

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type AOFBus struct {
	path string

	//format
	offsetSize int

	file   *os.File
	buffer *bufio.Writer
}

func NewAOFBus(path string, offsetSize int) (*AOFBus, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}
	buffer := bufio.NewWriter(file)

	return &AOFBus{
		path:       path,
		offsetSize: offsetSize,

		file:   file,
		buffer: buffer,
	}, nil
}

func (b AOFBus) EncodeLine(uid []byte, args [][]byte) []byte {
	fullCmd := bytes.Join(args, []byte(" "))
	return append(uid, fullCmd...)
}

func (b AOFBus) DecodeLine(line []byte) (offset []byte, args [][]byte) {
	offset = line[:b.offsetSize]
	args = bytes.Split(line[b.offsetSize:], []byte(" "))
	return offset, args
}

func (b *AOFBus) Append(args [][]byte) error {
	uid := NewUID()
	line := append(b.EncodeLine(uid.Bytes(), args), '\n')
	if _, err := b.buffer.Write(line); err != nil {
		return err
	}

	//TODO: performance
	return b.Flush()
}

func (b *AOFBus) Flush() error {
	return b.buffer.Flush()
}

func (b *AOFBus) Close() error {
	return b.file.Close()
}

func (b *AOFBus) Sync(ctx context.Context, writer io.Writer, offset []byte) error {
	aofFile, err := os.OpenFile(b.path, os.O_RDONLY, os.ModePerm)
	defer aofFile.Close()
	if err != nil {
		return err
	}
	rd := bufio.NewReader(aofFile)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			//TODO: care about isPrefix == true
			line, _, err := rd.ReadLine()
			if err == io.EOF {
				time.Sleep(time.Millisecond * 10)
				continue
			}
			if err != nil {
				return err
			}
			if len(line) < b.offsetSize {
				return errors.New(fmt.Sprintf("Invalid aof format: %b", line))
			}
			timestamp := line[:b.offsetSize]
			if offset != nil && bytes.Compare(timestamp, offset) < 0 {
				continue
			}

			if _, err := writer.Write(line); err != nil {
				return err
			}
		}
	}
}
