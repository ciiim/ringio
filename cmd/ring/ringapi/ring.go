package ringapi

import (
	"crypto/md5"
	"log/slog"

	"github.com/ciiim/cloudborad/chunkpool"
	"github.com/ciiim/cloudborad/ringio"
	"github.com/urfave/cli/v2"
)

var Ring *RingAPI

type RingAPI struct {
	*ringio.Ring
	chunkPool *chunkpool.ChunkPool
}

func md5Hash(data []byte) []byte {
	sum := md5.Sum(data)
	return sum[:]
}

func init() {

}

func InitRingAPI(flags *cli.Context) {
	config := &ringio.RingConfig{
		Name:         flags.String("hostname"),
		Port:         flags.Int("port"),
		Replica:      flags.Int("replica"),
		ChunkMaxSize: ringio.DefaultChunkSize,
		HashFn:       md5Hash,
		RootPath:     flags.String("root"),
		LogLevel:     slog.LevelInfo,
	}

	Ring = NewRingAPI(ringio.NewRing(config))

}

func NewRingAPI(ring *ringio.Ring) *RingAPI {
	return &RingAPI{
		Ring:      ring,
		chunkPool: chunkpool.NewChunkPool(ring.StorageSystem.Config().ChunkMaxSize),
	}
}
