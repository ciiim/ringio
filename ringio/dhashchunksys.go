package ringio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"

	"github.com/ciiim/cloudborad/chunkpool"
	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/storage/hashchunk"
)

var (
	ErrStoreWhenRecovering  = errors.New("store when recovering")
	ErrDeleteWhenRecovering = errors.New("delete when recovering")
)

// distribute file system
type DHashChunkSystem struct {
	localSys *hashchunk.HashChunkSystem

	pool *chunkpool.ChunkPool

	replicaService *replica.ReplicaServiceG[*hashchunk.HashChunkInfo]

	recover recoveringChunk

	remote *rpcHashClient
	ns     *node.NodeServiceRO

	l *slog.Logger
}

var _ IDHashChunkSystem = (*DHashChunkSystem)(nil)

func (d *DHashChunkSystem) local() hashchunk.IHashChunkSystem {
	return d.localSys
}

func (d *DHashChunkSystem) PickNode(key []byte) *node.Node {
	return d.ns.Pick(key)
}

func NewDHCS(config *hashchunk.Config, ns *node.NodeServiceRO, logger *slog.Logger) *DHashChunkSystem {
	d := &DHashChunkSystem{
		localSys: hashchunk.NewHashChunkSystem(config),

		pool: chunkpool.NewChunkPool(config.ChunkMaxSize),

		replicaService: replica.NewG[*hashchunk.HashChunkInfo](3, ns),

		ns: ns,

		l: logger,
	}

	if config.EnableReplica {
		d.recover.Finalize = func(key []byte, finalCount int64) {
			if finalCount <= 0 {
				_ = d.Delete(key)
			} else {
				_ = d.localSys.UpdateInfo(
					key,
					func(info *hashchunk.Info) {
						info.ChunkInfo.ChunkCount = finalCount
					},
				)
			}
		}
	}

	d.remote = newRPCHashClient(d.pool)

	d.setReplicaFunctions()
	return d
}

func (d *DHashChunkSystem) Get(key []byte) (chunk IDHashChunk, err error) {
	defer func() {
		if err != nil {
			d.l.Error("Get chunk", "error", err)
		}
	}()

	ni := d.PickNode(key)
	if ni == nil {
		return nil, node.ErrNodeNotFound
	}
	// get from local
	if ni.Equal(d.ns.Self()) {
		df, err := d.getLocally(key)
		if errors.Is(err, hashchunk.ErrFileNotFound) {
			return d.RecoverChunk(key)
		} else if errors.Is(err, hashchunk.ErrChunkInfoNotFound) {
			//TODO:恢复chunk信息
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

func (d *DHashChunkSystem) StoreBytes(key []byte, filename string, size int64, v []byte, extra *hashchunk.ExtraInfo) (err error) {

	defer func() {
		if err != nil {
			d.l.Error("Store bytes", "error", err)
		}
	}()

	ni := d.PickNode(key)
	if ni == nil {
		return node.ErrNodeNotFound
	}
	// store locally
	if ni.Equal(d.ns.Self()) {
		return d.storeBytesLocally(key, filename, size, v, extra)
	}

	reader := bytes.NewReader(v)

	// store remotely
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.put(ctx, ni, key, size, filename, reader)
}

func (d *DHashChunkSystem) StoreReader(key []byte, filename string, size int64, reader io.Reader, extra *hashchunk.ExtraInfo) (err error) {

	defer func() {
		if err != nil {
			d.l.Error("Store reader", "error", err)
		}
	}()

	ni := d.PickNode(key)
	if ni == nil {
		return node.ErrNodeNotFound
	}
	// store locally
	if ni.Equal(d.ns.Self()) {
		return d.storeReaderLocally(key, filename, size, reader, extra)
	}

	// store remotely
	log.Printf("[HashDFileSystem]Request redirect to %s.", ni.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.put(ctx, ni, key, size, filename, reader)
}

func (d *DHashChunkSystem) Delete(key []byte) (err error) {

	defer func() {
		if err != nil {
			d.l.Error("Delete", "error", err)
		}
	}()

	ni := d.PickNode(key)

	// no node
	if ni.Equal(nil) {
		return fmt.Errorf("no node for key %s", key)
	}

	// delete locally
	if ni.Equal(d.ns.Self()) {
		return d.deleteLocally(key)
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
	}, err
}

func (d *DHashChunkSystem) storeBytesLocally(key []byte, filename string, size int64, v []byte, extra *hashchunk.ExtraInfo) error {

	if d.recover.isRecovering(key) {
		// 假如isRecovering为真，但addCount失败，说明已经恢复完成，可以继续执行后面的函数
		if ok := d.recover.addCount(key); ok {
			return nil
		}
	}

	return d.local().StoreBytes(key, filename, size, v, extra)
}

func (d *DHashChunkSystem) storeReaderLocally(key []byte, filename string, size int64, reader io.Reader, extra *hashchunk.ExtraInfo) error {

	if d.recover.isRecovering(key) {
		// 假如isRecovering为真，但addCount失败，说明已经恢复完成，可以继续执行后面的函数
		if ok := d.recover.addCount(key); ok {
			return nil
		}
	}

	return d.local().StoreReader(key, filename, size, reader, extra)
}

func (d *DHashChunkSystem) deleteLocally(key []byte) error {

	if d.recover.isRecovering(key) {
		// 假如isRecovering为真，但minusCount失败，说明已经恢复完成，可以继续执行后面的函数
		if ok := d.recover.minusCount(key); ok {
			return nil
		}
	}

	return d.local().Delete(key)
}

func (d *DHashChunkSystem) Node() *node.Node {
	return d.ns.Self()
}

func (d *DHashChunkSystem) RecoverChunk(key []byte) (IDHashChunk, error) {
	if !d.Config().EnableReplica {
		return nil, hashchunk.ErrFileNotFound
	}

	// 防止重复恢复Chunk
	if ok := d.recover.registerChunk(key); !ok {
		// 如果已经在恢复中，直接返回
		// XXX: 可以让请求block在这里等待，直到恢复完成。
		// 目前先让请求者自己处理
		return nil, ErrChunkRecovering
	}
	defer d.recover.unregisterChunk(key)

	chunk, err := d.FindChunk(key)
	if err != nil {
		return nil, err
	}

	//先存到本地
	if err = func() error {
		writer, err := d.local().CreateChunk(key, chunk.Info().ChunkInfo.ChunkName, chunk.Info().ChunkInfo.Size(), chunk.Info().ExtraInfo)
		if err != nil {
			return err
		}
		defer writer.Close()
		if _, err = io.Copy(writer, chunk); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return nil, err
	}

	//再返给调用者
	if _, err = chunk.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return chunk, nil

}

func (d *DHashChunkSystem) FindChunk(key []byte) (IDHashChunk, error) {

	reader, replicaInfo, err := d.replicaService.RecoverReplica(key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	hashChunk := &hashchunk.HashChunk{
		ReadSeekCloser: reader,
	}

	//chunk信息
	chunkInfo := replicaInfo.Custom
	replicaInfo.ClearCustom()

	//还原包含chunk和replica信息的Info
	info := hashchunk.NewInfo(chunkInfo, hashchunk.NewExtraInfo("replica", replicaInfo))

	hashChunk.SetInfo(info)

	return &DHashChunk{
		HashChunk: hashChunk,
	}, nil
}

func (d *DHashChunkSystem) Config() *hashchunk.Config {
	return d.local().Config()
}
