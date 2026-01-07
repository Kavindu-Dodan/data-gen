package internal

import "bytes"

// trackedBuffer wraps a buffer and tracks its size to enable batch size limits.
type trackedBuffer struct {
	buf bytes.Buffer
	len int64
}

func newTrackedBuffer() trackedBuffer {
	return trackedBuffer{
		buf: bytes.Buffer{},
	}
}

func (t *trackedBuffer) write(bytes []byte) error {
	t.len += int64(len(bytes))
	_, err := t.buf.Write(bytes)
	return err
}

func (t *trackedBuffer) getAndReset() []byte {
	b := make([]byte, t.buf.Len())
	copy(b, t.buf.Bytes())
	t.buf.Reset()
	t.len = 0
	return b
}

func (t *trackedBuffer) size() int64 {
	return t.len
}
