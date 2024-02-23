package hashchunk

import (
	"io"
	"os"
)

var (
	ErrCannotBeWritten = os.ErrPermission
)

type HashChunk struct {
	io.ReadSeekCloser
	info *Info
}

func (c *HashChunk) Info() *Info {
	return c.info
}

func (c *HashChunk) SetInfo(info *Info) {
	c.info = info
}

type HashChunkWriteCloser struct {
	io.WriteCloser
}

func warpHashChunkWriteCloser(wc io.WriteCloser) HashChunkWriteCloser {
	return HashChunkWriteCloser{
		wc,
	}
}
