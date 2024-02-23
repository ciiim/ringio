package chunkpool

import (
	"sync"
)

type ChunkPool struct {
	pool      sync.Pool
	chunkSize int64
}

func NewChunkPool(chunkSize int64) *ChunkPool {
	cp := &ChunkPool{
		pool:      sync.Pool{},
		chunkSize: chunkSize,
	}
	cp.pool.New = func() any {
		return NewChunkBuffer(cp.chunkSize)
	}
	return cp
}

func (c *ChunkPool) Get() *ChunkBuffer {
	return c.pool.Get().(*ChunkBuffer)
}

func (c *ChunkPool) Put(cb *ChunkBuffer) {
	clear(cb.buffer)
	c.pool.Put(cb)
}
