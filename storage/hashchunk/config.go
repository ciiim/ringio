package hashchunk

type Config struct {
	chunkMaxSize int64
	hashFn       Hash
}

func (c *Config) ChunkMaxSize() int64 {
	return c.chunkMaxSize
}

func (c *Config) HashFn() Hash {
	return c.hashFn
}
