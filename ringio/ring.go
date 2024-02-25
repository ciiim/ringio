package ringio

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"github.com/ciiim/cloudborad/storage/tree"
	"github.com/ciiim/cloudborad/storage/types"
)

const (
	DefaultChunkSize types.Byte = 1024 * 1024 * 32 // 32MB
	DefaultPort      int        = 9631
	DefualtReplica   int        = 20
	DefaultName      string     = "ring"
)

type Ring struct {
	StorageSystem IDHashChunkSystem
	FrontSystem   ITreeDFileSystem

	config *RingConfig

	ringName string

	nodeService *node.NodeService

	rpcServer *rpcServer

	l *slog.Logger
}

func NewRing(config *RingConfig) *Ring {

	node := node.NewNodeService(config.Name, config.Port, config.Replica)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: config.LogLevel}))

	storageSystem := NewDHCS(
		&DHCSConfig{
			HCSConfig: &hashchunk.Config{
				RootPath:          filepath.Join(config.RootPath, "ds"),
				Capacity:          -1,
				ChunkMaxSize:      config.ChunkMaxSize,
				HashFn:            config.HashFn,
				CalcStoragePathFn: nil,
			},
			EnableReplica: config.EnableReplica,
		},
		node.NodeServiceRO(),
		logger,
	)

	frontSystem := NewDTreeFileSystem(
		&tree.Config{
			RootPath: filepath.Join(config.RootPath, "tfs"),
		},
		node.NodeServiceRO(),
		logger,
	)

	ring := &Ring{
		ringName: config.Name,

		// 节点服务
		nodeService: node,

		// 存储系统
		StorageSystem: storageSystem,

		// 文件系统
		FrontSystem: frontSystem,

		// rpc服务
		rpcServer: newRPCFSServer(node.NodeServiceRO(), storageSystem, frontSystem),

		// 日志
		l: logger,

		config: config,
	}
	return ring
}

func (r *Ring) Serve() {
	if r.nodeService == nil {
		r.l.Error("[Ring] No node service found")
		return
	}
	go func() {
		_ = r.nodeService.Run()
	}()
	rpcPort, _ := strconv.ParseInt(r.nodeService.Self().Port(), 10, 64)

	r.l.Info("[Ring] Service serve on", "port", rpcPort)
	r.rpcServer.serve(strconv.FormatInt(rpcPort, 10))
}

func (r *Ring) Close() error {
	r.nodeService.Shutdown()
	return nil
}
