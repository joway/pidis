package util

import "sync"

type StreamBus struct {
	mutex sync.Mutex
	ch    chan []byte
}

func NewStreamBus() *StreamBus {
	return &StreamBus{ch: make(chan []byte, 1024)}
}

func (w *StreamBus) Read() <-chan []byte {
	return w.ch
}

func (w *StreamBus) Write(buf []byte) (int, error) {
	w.mutex.Lock()
	w.ch <- buf
	w.mutex.Unlock()
	return len(buf), nil
}

func (w *StreamBus) Close() error {
	close(w.ch)
	return nil
}
