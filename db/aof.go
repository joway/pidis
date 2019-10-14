package db

import (
	"bufio"
	"bytes"
	"github.com/rs/xid"
)

type AOFWriter struct {
	buffer *bufio.Writer
}

func encode(uuid []byte, cmd [][]byte) []byte {
	fullCmd := bytes.Join(cmd, []byte(" "))
	return append(append(uuid, fullCmd...), '\n')
}

func NewAOFWriter(writer *bufio.Writer) *AOFWriter {
	return &AOFWriter{
		buffer: writer,
	}
}

func (w *AOFWriter) Append(cmd [][]byte) error {
	guid := xid.New()
	line := encode(guid.Bytes(), cmd)
	if _, err := w.buffer.Write(line); err != nil {
		return err
	}
	return w.Flush()
}

func (w *AOFWriter) Flush() error {
	return w.buffer.Flush()
}
