package hashchunk

type Config struct {
	RootPath          string
	Capacity          int64
	ChunkMaxSize      int64
	HashFn            Hash
	CalcStoragePathFn CalcChunkStoragePathFn

	// 是否启用副本
	EnableReplica bool
}
