package hashchunk

import (
	"bufio"
	"io"
)

type HashChunk struct {
	io.ReadCloser
	info *HashChunkInfo
}

func (c *HashChunk) Info() IHashChunkInfo {
	return c.info
}

func (c *HashChunk) SetInfo(info *HashChunkInfo) {
	c.info = info
}

type HashChunkWriteCloser struct {
	closed bool
	w      *bufio.Writer
	closer io.Closer
}

func warpHashChunkWriteCloser(wc io.WriteCloser) *HashChunkWriteCloser {
	bufw := bufio.NewWriter(wc)
	return &HashChunkWriteCloser{
		w:      bufw,
		closer: wc,
	}
}

func (c *HashChunkWriteCloser) Flush() error {
	return c.w.Flush()
}

func (c *HashChunkWriteCloser) Close() error {
	if err := c.closer.Close(); err != nil {
		return err
	}
	c.closed = true
	return nil
}

func (c *HashChunkWriteCloser) Write(p []byte) (n int, err error) {
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	return c.w.Write(p)
}
