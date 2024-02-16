package chunkpool

import (
	"sync"
)

type ChunkPool struct {
	mu        sync.Mutex
	pool      sync.Pool
	chunkSize int64
}

func NewChunkPool(chunkSize int64) *ChunkPool {
	cp := &ChunkPool{
		mu:        sync.Mutex{},
		pool:      sync.Pool{},
		chunkSize: chunkSize,
	}
	cp.pool.New = func() any {
		return NewChunkBuffer(cp.chunkSize)
	}
	return cp
}

func (c *ChunkPool) Get() *ChunkBuffer {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.pool.Get().(*ChunkBuffer)
}

func (c *ChunkPool) Put(cb *ChunkBuffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	clear(cb.buffer)
	c.pool.Put(cb)
}
