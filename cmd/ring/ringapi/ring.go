package ringapi

import (
	"crypto/md5"
	"log/slog"
	"os"

	"github.com/ciiim/cloudborad/chunkpool"
	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/ringio"
)

var Ring *RingAPI

type RingAPI struct {
	ring      *ringio.Ring
	chunkPool *chunkpool.ChunkPool
}

func md5Hash(data []byte) []byte {
	sum := md5.Sum(data)
	return sum[:]
}

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ringio.DefaultName
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	nodeService := node.NewNodeService(hostname, ringio.DefaultPort, ringio.DefualtReplica)
	nodesro := nodeService.NodeServiceRO()
	tfs := ringio.NewDTreeFileSystem("./ring/fs", nodesro)
	hcs := ringio.NewDHCS("./ring/storage", -1, ringio.DefaultChunkSize, nodesro, md5Hash, nil)

	Ring = NewRingAPI(ringio.NewRing(hostname, logger, nodeService, tfs, hcs))
}

func NewRingAPI(ring *ringio.Ring) *RingAPI {
	return &RingAPI{
		ring:      ring,
		chunkPool: chunkpool.NewChunkPool(ring.StorageSystem.Config().ChunkMaxSize()),
	}
}
