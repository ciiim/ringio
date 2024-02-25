package ringio

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
)

var (
	ErrChunkRecovering = errors.New("chunk recovering")
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

	// 文件恢复完成后的回调
	Finalize func(key []byte, finalCount int64)
}

// count记录恢复过程中的操作次数，+1表示有一个存储操作，-1表示有一个删除操作
// 如果最后恢复完的chunkinfo的引用计数减去count小于等于0，删除该chunk
type count struct {
	count int64
}

func (r *recoveringChunk) isRecovering(key []byte) bool {
	k := string(key)
	_, ok := r.m.Load(k)
	return ok
}

// 并发安全
func (r *recoveringChunk) registerChunk(key []byte) bool {
	k := string(key)
	_, exist := r.m.LoadOrStore(k, &count{0})
	if exist {
		return false
	}
	return true
}

func (r *recoveringChunk) unregisterChunk(key []byte) {
	k := string(key)
	v, ok := r.m.LoadAndDelete(k)
	if !ok {
		return
	}

	count := v.(*count)

	if r.Finalize != nil {
		r.Finalize(key, count.count)
	}
}

func (r *recoveringChunk) addCount(key []byte) (ok bool) {
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

func (r *recoveringChunk) minusCount(key []byte) (ok bool) {
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

func (d *DHashChunkSystem) setReplicaFunctions() {
	d.replicaService.SetFunctions(d.putReplica, d.getReplica, d.delReplica, d.checkReplica, d.updateReplicaInfo)
}

func (d *DHashChunkSystem) putReplica(
	nodeID string,
	reader io.Reader,
	info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo],
) error {
	n := d.ns.GetByNodeID(nodeID)
	if n == nil {
		return replica.ErrNoReplicaNode
	}

	if n.Equal(d.ns.Self()) {
		return node.ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	return d.remote.putReplica(ctx, n, reader, info)

}

func (d *DHashChunkSystem) getReplica(
	nodeID string,
	key []byte,
) (io.ReadSeekCloser, *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo], error) {
	n := d.ns.GetByNodeID(nodeID)
	if n == nil {
		return nil, nil, replica.ErrNoReplicaNode
	}

	if n.Equal(d.ns.Self()) {
		return nil, nil, node.ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	return d.remote.getReplica(ctx, n, &fspb.Key{Key: key})
}

func (d *DHashChunkSystem) delReplica(
	nodeID string,
	key []byte,
) error {
	n := d.ns.GetByNodeID(nodeID)
	if n == nil {
		return replica.ErrNoReplicaNode
	}

	if n.Equal(d.ns.Self()) {
		return node.ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	return d.remote.delReplica(ctx, n, &fspb.Key{Key: key})
}

func (d *DHashChunkSystem) checkReplica(
	nodeID string,
	info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo],
) error {
	n := d.ns.GetByNodeID(nodeID)
	if n == nil {
		return replica.ErrNoReplicaNode
	}

	if n.Equal(d.ns.Self()) {
		return node.ErrSelfNode
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	return d.remote.checkReplica(ctx, n, info)
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
