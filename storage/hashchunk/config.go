package hashchunk

type Config struct {
	RootPath          string
	Capacity          int64
	ChunkMaxSize      int64
	HashFn            Hash
	CalcStoragePathFn CalcChunkStoragePathFn
}
