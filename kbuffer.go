package sp2p

import (
	"bytes"
	"sync"
)

func NewKBuffer(delim []byte) *KBuffer {
	return &KBuffer{delim: delim}
}

type KBuffer struct {
	buf   []byte
	delim []byte
	sync.RWMutex
}

func (t *KBuffer) Next(b []byte) [][]byte {
	t.Lock()
	defer t.Unlock()

	if b == nil {
		return nil
	}

	t.buf = append(t.buf, b...)

	if len(t.buf) < 1 {
		return nil
	}

	if !bytes.Contains(t.buf, t.delim) {
		return nil
	}

	d := bytes.Split(t.buf, t.delim)
	if len(d) < 1 {
		return nil
	}

	t.buf = d[len(d)-1]
	return d[:len(d)-1]
}
