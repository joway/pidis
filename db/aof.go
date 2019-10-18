package db

import (
	"bufio"
	"bytes"
	"context"
	"github.com/joway/pikv/types"
	"github.com/tidwall/redcon"
	"io"
	"os"
	"sync"
	"time"
)

type AOFBus struct {
	lock *sync.Mutex
	path string

	//format
	offsetSize int

	file   *os.File
	buffer *bufio.Writer
}

func EncodeAOF(uid []byte, args [][]byte) []byte {
	var encoded []byte
	encoded = redcon.AppendArray(encoded, len(args)+1)
	encoded = redcon.AppendBulk(encoded, uid)
	for _, arg := range args {
		encoded = redcon.AppendBulk(encoded, arg)
	}
	return encoded
}

func DecodeAOF(content []byte) (uid []byte, args [][]byte, leftover []byte, err error) {
	isCompleted, args, _, leftover, err := redcon.ReadNextCommand(content, nil)
	if err != nil {
		return nil, nil, content, types.ErrInvalidAOFFormat
	}
	if !isCompleted {
		return nil, nil, content, nil
	}
	return args[0], args[1:], leftover, nil
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

		lock: &sync.Mutex{},
	}, nil
}

func (b *AOFBus) Append(args [][]byte) error {
	uid := NewUID()
	line := EncodeAOF(uid.Bytes(), args)
	b.lock.Lock()
	defer b.lock.Unlock()
	if _, err := b.buffer.Write(line); err != nil {
		return err
	}
	return nil
}

func (b *AOFBus) Flush() error {
	b.lock.Lock()
	defer b.lock.Unlock()
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

	//1. Find the position by offset
	var buffer []byte
	//TODO: tuning
	buf := make([]byte, 1024)
	var packet []byte
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			size, err := rd.Read(buf)
			if err == io.EOF || size == 0 {
				//TODO: tuning
				time.Sleep(time.Millisecond * 10)
				continue
			}
			if err != nil {
				return err
			}

			buffer = append(buffer, buf[:size]...)
			//parser sections
			for {
				uid, args, leftover, err := DecodeAOF(buffer)
				if err != nil {
					return err
				}
				//uncompleted buffer
				if uid == nil && args == nil {
					break
				}
				if bytes.Compare(uid, offset) < 0 {
					buffer = leftover
					//skip
					continue
				}
				packet = append(packet, buffer[:len(buffer)-len(leftover)]...)
				buffer = leftover
			}

			if packet != nil {
				if _, err := writer.Write(packet); err != nil {
					return err
				}
				packet = nil
			}
		}
	}
}
