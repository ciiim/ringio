package chunkpool

import (
	"bytes"
	"errors"
)

var (
	FullBuffer = errors.New("buffer is full")
)

type ChunkBuffer struct {
	buffer       []byte
	bufferLength int
	buffered     int
}

func NewChunkBuffer(size int64) *ChunkBuffer {
	buffer := make([]byte, size)
	return &ChunkBuffer{
		buffer:       buffer,
		bufferLength: len(buffer),
		buffered:     0,
	}
}

func (c *ChunkBuffer) Hash(h func(s []byte) []byte) []byte {
	return h(c.buffer[:c.buffered])
}

func (c *ChunkBuffer) Reset() {
	c.buffered = 0

}

func (c *ChunkBuffer) Write(b []byte) (int, error) {
	n := len(b)
	if n == 0 {
		return 0, nil
	}
	if n > c.bufferLength-c.buffered {
		return 0, FullBuffer
	}
	copy(c.buffer[c.buffered:], b)
	c.buffered += n
	return n, nil
}

type ChunkCloser struct {
	*bytes.Reader
	CloseFn func()
}

func NewChunkCloser(r *bytes.Reader, closeFn func()) *ChunkCloser {
	return &ChunkCloser{
		Reader:  r,
		CloseFn: closeFn,
	}
}

func (cc *ChunkCloser) Close() error {
	cc.CloseFn()
	return nil
}

func (c *ChunkBuffer) ReadCloser(pool *ChunkPool) *ChunkCloser {
	return NewChunkCloser(bytes.NewReader(c.buffer), func() {
		c.putSelf(pool)()
	})
}

func (c *ChunkBuffer) putSelf(pool *ChunkPool) func() {
	return func() {
		pool.Put(c)
	}
}
