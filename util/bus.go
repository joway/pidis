package util

type StreamBus struct {
	ch chan []byte
}

func NewStreamBus() *StreamBus {
	return &StreamBus{make(chan []byte, 1024)}
}

func (w *StreamBus) Read() <-chan []byte {
	return w.ch
}

func (w *StreamBus) Write(buf []byte) (int, error) {
	w.ch <- buf
	return len(buf), nil
}

func (w *StreamBus) Close() error {
	close(w.ch)
	return nil
}
