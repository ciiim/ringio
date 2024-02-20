package ringio

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
)

/*
恢复中的Chunk的外部操作

  - 存储
    如果在一个chunk恢复中，有存储操作进来，先被记录下来，不会再执行一次恢复操作。当恢复完成后，执行存储操作。
  - 删除
    如果在一个chunk恢复中，有删除操作进来，先被记录下来。当恢复完成后，执行删除操作。
*/
type recoveringChunk struct {
	// map[<[]byte>] *count <- int64
	m sync.Map
}

type count struct {
	count int64
}

func (r *recoveringChunk) registerChunk(key []byte) {
	k := string(key)
	r.m.Store(k, &count{0})
}

func (r *recoveringChunk) unregisterChunk(key []byte) {
	k := string(key)
	r.m.Delete(k)
}

func (r *recoveringChunk) addAction(key []byte) (ok bool) {
	k := string(key)
	v, ok := r.m.Load(k)
	if !ok {
		return false
	}

	count := v.(*count)

	atomic.AddInt64(&count.count, 1)

	r.m.Store(k, count)
	return true
}

func (r *recoveringChunk) delAction(key []byte) (ok bool) {
	k := string(key)
	v, ok := r.m.Load(k)
	if !ok {
		return false
	}
	count := v.(*count)

	atomic.AddInt64(&count.count, -1)

	r.m.Store(k, count)

	return true
}

var (
	ErrSelfNode = errors.New("self node")
)

func (d *DHashChunkSystem) setReplicaFunctions() {
	d.replicaService.SetFunctions(d.putReplica, d.getReplica, d.delReplica, d.checkReplica, d.updateReplicaInfo)
}

func (d *DHashChunkSystem) putReplica(
	nodeID string,
	reader io.Reader,
	info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo],
) error {
	node := d.ns.GetByNodeID(nodeID)
	if node == nil {
		return replica.ErrNoReplicaNode
	}

	if node.Equal(d.ns.Self()) {
		return ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	chunkInfo := info.Custom

	return d.remote.putReplica(ctx, node, reader, chunkInfo, info)

}

func (d *DHashChunkSystem) getReplica(
	nodeID string,
	key []byte,
) (io.ReadSeekCloser, *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo], error) {
	node := d.ns.GetByNodeID(nodeID)
	if node == nil {
		return nil, nil, replica.ErrNoReplicaNode
	}

	if node.Equal(d.ns.Self()) {
		return nil, nil, ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	return d.remote.getReplica(ctx, node, &fspb.Key{Key: key})
}

func (d *DHashChunkSystem) delReplica(
	nodeID string,
	key []byte,
) error {
	node := d.ns.GetByNodeID(nodeID)
	if node == nil {
		return replica.ErrNoReplicaNode
	}

	if node.Equal(d.ns.Self()) {
		return ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	return d.remote.delReplica(ctx, node, &fspb.Key{Key: key})
}

func (d *DHashChunkSystem) checkReplica(
	nodeID string,
	info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo],
) error {
	node := d.ns.GetByNodeID(nodeID)
	if node == nil {
		return replica.ErrNoReplicaNode
	}

	if node.Equal(d.ns.Self()) {
		return ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	return d.remote.checkReplica(ctx, node, info)
}

func (d *DHashChunkSystem) updateReplicaInfo(
	nodeID string,
	info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo],
) error {
	node := d.ns.GetByNodeID(nodeID)
	if node == nil {
		return nil
	}
	if node.Equal(d.ns.Self()) {
		return nil
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	_ = d.remote.updateReplicaInfo(ctx, node, info)

	return nil
}
