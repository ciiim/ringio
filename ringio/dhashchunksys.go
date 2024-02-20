package ringio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/ciiim/cloudborad/chunkpool"
	dlogger "github.com/ciiim/cloudborad/debug"
	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/storage/hashchunk"
)

// distribute file system
type DHashChunkSystem struct {
	localSys *hashchunk.HashChunkSystem

	pool *chunkpool.ChunkPool

	replicaService *replica.ReplicaServiceG[*hashchunk.HashChunkInfo]

	recoveringChunk map[string]struct{}

	remote *rpcHashClient
	ns     *node.NodeServiceRO
}

var _ IDHashChunkSystem = (*DHashChunkSystem)(nil)

func (d *DHashChunkSystem) local() hashchunk.IHashChunkSystem {
	return d.localSys
}

func (d *DHashChunkSystem) PickNode(key []byte) *node.Node {
	return d.ns.Pick(key)
}

func NewDHCS(rootPath string, capacity int64, chunkSize int64, ns *node.NodeServiceRO, hashFn hashchunk.Hash, calcStoragePathFn hashchunk.CalcChunkStoragePathFn) *DHashChunkSystem {
	d := &DHashChunkSystem{
		localSys: hashchunk.NewHashChunkSystem(rootPath, capacity, chunkSize, hashFn, calcStoragePathFn),

		pool: chunkpool.NewChunkPool(chunkSize),

		replicaService: replica.NewG[*hashchunk.HashChunkInfo](3, ns),

		ns: ns,
	}

	d.remote = newRPCHashClient(d.pool)

	d.setReplicaFunctions()
	return d
}

func (d *DHashChunkSystem) Get(key []byte) (IDHashChunk, error) {
	dlogger.Dlog.LogDebugf("[HashDFileSystem]", "Get by key '%s'", key)
	ni := d.PickNode(key)
	if ni == nil {
		return nil, node.ErrNodeNotFound
	}
	// get from local
	if ni.Equal(d.ns.Self()) {
		df, err := d.getLocally(key)
		if errors.Is(err, hashchunk.ErrFileNotFound) {
			return d.HealChunk(key)
		} else {
			return df, err
		}
	}

	// get from remote
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()

	resp, err := d.remote.get(ctx, ni, key)
	return resp, err
}

func (d *DHashChunkSystem) StoreBytes(key []byte, filename string, v []byte, extra *hashchunk.ExtraInfo) error {
	dlogger.Dlog.LogDebugf("[HashDFileSystem]", "Store by key '%s', name '%s'", key, filename)
	ni := d.PickNode(key)
	if ni == nil {
		return node.ErrNodeNotFound
	}
	// store locally
	if ni.Equal(d.ns.Self()) {
		return d.local().StoreBytes(key, filename, v, extra)
	}

	reader := bytes.NewReader(v)

	// store remotely
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.put(ctx, ni, key, filename, reader)
}

func (d *DHashChunkSystem) StoreReader(key []byte, filename string, reader io.Reader, extra *hashchunk.ExtraInfo) error {
	dlogger.Dlog.LogDebugf("[HashDFileSystem]", "Store by key '%s', name '%s'", key, filename)
	ni := d.PickNode(key)
	if ni == nil {
		return node.ErrNodeNotFound
	}
	// store locally
	if ni.Equal(d.ns.Self()) {
		return d.storeLocally(key, filename, reader, extra)
	}

	// store remotely
	log.Printf("[HashDFileSystem]Request redirect to %s.", ni.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.put(ctx, ni, key, filename, reader)
}

func (d *DHashChunkSystem) Delete(key []byte) error {
	ni := d.PickNode(key)
	if ni == nil {
		return node.ErrNodeNotFound
	}
	// delete locally
	if ni.Equal(d.ns.Self()) {
		return d.deleteLocally(key)
	}

	// no node
	if ni.Equal(nil) {
		return fmt.Errorf("no node for key %s", key)
	}

	// delete remotely
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.delete(ctx, ni, key)
}

func (d *DHashChunkSystem) getLocally(key []byte) (*DHashChunk, error) {
	chunk, err := d.local().Get(key)
	return &DHashChunk{
			HashChunk: chunk,
		},
		err
}

func (d *DHashChunkSystem) storeLocally(key []byte, filename string, reader io.Reader, extra *hashchunk.ExtraInfo) error {
	return d.local().StoreReader(key, filename, reader, extra)
}

func (d *DHashChunkSystem) deleteLocally(key []byte) error {
	return d.local().Delete(key)
}

func (d *DHashChunkSystem) Node() *node.Node {
	return d.ns.Self()
}

func (d *DHashChunkSystem) HealChunk(key []byte) (IDHashChunk, error) {

	reader, replicaInfo, err := d.replicaService.RecoverReplica(key)
	if err != nil {
		return nil, err
	}

	hashChunk := new(hashchunk.HashChunk)

	hashChunk.ReadCloser = reader

	chunkInfo := replicaInfo.Custom

	// 把chunkInfo单独拿出来，后续replicaInfo存盘的时候就不会把chunkinfo再存一次
	replicaInfo.Custom = nil

	info := hashchunk.NewInfo(chunkInfo, hashchunk.NewExtraInfo("replica", replicaInfo))

	hashChunk.SetInfo(info)

	return &DHashChunk{
		HashChunk: hashChunk,
	}, nil
}

func (d *DHashChunkSystem) Config() *hashchunk.Config {
	return d.local().Config()
}
