package ringio

import (
	"log"
	"log/slog"
	"strconv"

	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/storage/types"
)

const (
	DefaultChunkSize types.Byte = 1024 * 1024 * 32 // 32MB
	DefaultPort      int        = 9631
	DefualtReplica   int        = 5
	DefaultName      string     = "ring"
)

type Ring struct {
	ringName string

	nodeService *node.NodeService

	rpcServer *rpcServer

	FrontSystem ITreeDFileSystem

	StorageSystem IDHashChunkSystem

	l *slog.Logger
}

func NewRing(ringName string, logger *slog.Logger, nodeService *node.NodeService, frontSystem ITreeDFileSystem, storageSystem IDHashChunkSystem) *Ring {
	if frontSystem == nil {
		log.Println("[Ring init] Empty front system")
		return nil
	}
	if storageSystem == nil {
		log.Println("[Ring init] Empty store system")
		return nil
	}

	ring := &Ring{
		ringName: ringName,

		// 节点服务
		nodeService: nodeService,

		// 存储系统
		StorageSystem: storageSystem,

		// 文件系统
		FrontSystem: frontSystem,

		// rpc服务
		rpcServer: newRPCFSServer(nodeService.NodeServiceRO(), storageSystem, frontSystem),

		// 日志
		l: logger,
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
	//FIXME: 要使用同一个端口
	rpcPort, _ := strconv.ParseInt(r.nodeService.Self().Port(), 10, 64)
	rpcPort++

	r.l.Info("[Ring] node service serve on", "port", r.nodeService.Self().Port())
	r.rpcServer.serve(strconv.FormatInt(rpcPort, 10))
}

func (r *Ring) Close() error {
	r.nodeService.Shutdown()
	return nil
}
