package bytespool

import (
	"bytes"
	"sync"
)

var (
	pool *sync.Pool
)

func init() {
	pool.New = func() interface{} {
		return &bytes.Buffer{}
	}
}

func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func Put(b *bytes.Buffer) {
	pool.Put(b)
}
