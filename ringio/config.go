package ringio

type RingConfig struct {
	Name      string
	Port      int
	Replica   int
	ChunkSize int

	// TODO: 是否开启副本备份特性
	EnableReplica bool

	// TODO: 是否开启数据校验特性
	EnableDataCheck bool

	// TODO: 是否开启数据压缩特性
	EnableDataCompress bool

	// TODO: 是否开启数据加密特性
	EnableDataEncrypt bool

	// TODO: 是否开启小文件优化特性
	EnableSmallFileOptimize bool
}
