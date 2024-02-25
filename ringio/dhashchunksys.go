package ringio

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/ciiim/cloudborad/chunkpool"
	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"github.com/ciiim/cloudborad/util"
)

var (
	ErrStoreWhenRecovering  = errors.New("store when recovering")
	ErrDeleteWhenRecovering = errors.New("delete when recovering")
)

type HashChunkReplicaInfo = replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo]

type DHCSConfig struct {
	HCSConfig     *hashchunk.Config
	EnableReplica bool
}

// distribute file system
type DHashChunkSystem struct {
	localSys *hashchunk.HashChunkSystem

	pool *chunkpool.ChunkPool

	replicaService *replica.ReplicaServiceG[*hashchunk.HashChunkInfo]

	recover recoveringChunk

	remote *rpcHashClient
	ns     *node.NodeServiceRO

	config *DHCSConfig

	l *slog.Logger
}

var _ IDHashChunkSystem = (*DHashChunkSystem)(nil)

func (d *DHashChunkSystem) local() hashchunk.IHashChunkSystem {
	return d.localSys
}

func (d *DHashChunkSystem) PickNode(key []byte) *node.Node {
	return d.ns.Pick(key)
}

func NewDHCS(config *DHCSConfig, ns *node.NodeServiceRO, logger *slog.Logger) *DHashChunkSystem {
	d := &DHashChunkSystem{
		localSys: hashchunk.NewHashChunkSystem(config.HCSConfig),

		pool: chunkpool.NewChunkPool(config.HCSConfig.ChunkMaxSize),

		ns: ns,

		config: config,

		l: logger,
	}

	d.remote = newRPCHashClient(d.pool)

	if config.EnableReplica {
		d.replicaService = replica.NewG[*hashchunk.HashChunkInfo](3, ns)

		d.recover.Finalize = func(key []byte, finalCount int64) {
			info, err := d.local().GetInfo(key)
			if err != nil {
				return
			}
			if info.ChunkInfo.ChunkCount+finalCount <= 0 {
				_ = d.Delete(key)
			} else {
				_ = d.local().UpdateInfo(
					key,
					func(info *hashchunk.Info) {
						info.ChunkInfo.ChunkCount += finalCount
					},
				)
			}
		}
		d.setReplicaFunctions()
		d.l.Info("Replica enabled")
	}

	return d
}

func (d *DHashChunkSystem) Get(key []byte) (chunk IDHashChunk, err error) {
	defer func(err *error) {
		if panic := recover(); panic != nil {
			d.l.Error("Get chunk", "panic", panic)
		}
		if *err != nil {
			d.l.Error("Get chunk", "error", *err)
		}
	}(&err)

	ni := d.PickNode(key)
	if ni == nil {
		return nil, node.ErrNodeNotFound
	}
	// get from local
	if ni.Equal(d.ns.Self()) {
		return d.GetLocally(key)
	}

	// get from remote
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()

	resp, err := d.remote.get(ctx, ni, key)

	return resp, err
}

func (d *DHashChunkSystem) StoreBytes(key []byte, filename string, size int64, v []byte, extra *hashchunk.ExtraInfo) (err error) {

	defer func(err *error) {
		if err != nil {
			d.l.Error("Store bytes", "error", err)
		}
	}(&err)

	ni := d.PickNode(key)
	if ni == nil {
		return node.ErrNodeNotFound
	}
	// store locally
	if ni.Equal(d.ns.Self()) {
		err := d.StoreBytesLocally(key, filename, size, v, extra)
		return err
	}

	reader := bytes.NewReader(v)

	// store remotely
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.put(ctx, ni, key, size, filename, reader)
}

func (d *DHashChunkSystem) StoreReader(key []byte, filename string, size int64, reader io.Reader, extra *hashchunk.ExtraInfo) (err error) {

	defer func(err *error) {
		if *err != nil {
			d.l.Error("Store reader", "error", err)
		}
	}(&err)

	ni := d.PickNode(key)
	if ni == nil {
		return node.ErrNodeNotFound
	}
	// store locally
	if ni.Equal(d.ns.Self()) {
		err := d.StoreReaderLocally(key, filename, size, reader, extra)
		return err
	}

	// store remotely
	d.l.Info("Store reader", "remote", ni.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.put(ctx, ni, key, size, filename, reader)
}

func (d *DHashChunkSystem) Delete(key []byte) (err error) {

	defer func(err *error) {
		if *err != nil {
			d.l.Error("Delete", "error", err)
		}
	}(&err)

	ni := d.PickNode(key)

	// no node
	if ni.Equal(nil) {
		return fmt.Errorf("no node for key %s", key)
	}

	// delete locally
	if ni.Equal(d.ns.Self()) {
		err := d.DeleteLocally(key)
		return err
	}

	// delete remotely
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.delete(ctx, ni, key)
}

func (d *DHashChunkSystem) GetLocally(key []byte) (IDHashChunk, error) {
	chunk, err := d.local().Get(key)
	fmt.Printf("GetLocally %x err: %w\n", key, err)
	if errors.Is(err, hashchunk.ErrChunkNotFound) {
		return d.RecoverChunk(key)
	} else if errors.Is(err, hashchunk.ErrChunkInfoNotFound) {
		//TODO:恢复chunk信息
	}
	return &DHashChunk{
		HashChunk: chunk,
	}, err
}

func (d *DHashChunkSystem) StoreBytesLocally(key []byte, filename string, size int64, v []byte, extra *hashchunk.ExtraInfo) error {

	if d.recover.isRecovering(key) {
		// 假如isRecovering为真，但addCount失败，说明已经恢复完成，可以继续执行后面的函数
		if ok := d.recover.addCount(key); ok {
			return nil
		}
	}
	err := d.local().StoreBytes(key, filename, size, v, extra)
	if err != nil {
		return err
	}

	d.doReplicate(key)

	return nil
}

type warpCloserFn struct {
	io.WriteCloser
	close func()
}

func newWarpCloserFn(wc io.WriteCloser, close func()) io.WriteCloser {
	return warpCloserFn{
		WriteCloser: wc,
		close:       close,
	}
}

func (w warpCloserFn) Close() error {
	err := w.WriteCloser.Close()
	w.close()
	return err
}

func (d *DHashChunkSystem) CreateChunkLocally(key []byte, name string, size int64, extra *hashchunk.ExtraInfo) (io.WriteCloser, error) {

	if d.config.EnableReplica {
		if d.recover.isRecovering(key) {
			// 假如isRecovering为真，但addCount失败，说明已经恢复完成，可以继续执行后面的函数
			if ok := d.recover.addCount(key); ok {
				return nil, nil
			}
		}
	}

	wc, err := d.local().CreateChunk(key, name, size, extra)
	if err != nil {
		return nil, err
	}

	wc = newWarpCloserFn(wc, func() {
		d.doReplicate(key)
	})

	return wc, nil
}

func (d *DHashChunkSystem) StoreReaderLocally(key []byte, filename string, size int64, reader io.Reader, extra *hashchunk.ExtraInfo) error {

	if d.config.EnableReplica {
		if d.recover.isRecovering(key) {
			// 假如isRecovering为真，但addCount失败，说明已经恢复完成，可以继续执行后面的函数
			if ok := d.recover.addCount(key); ok {
				return nil
			}
		}
	}
	err := d.local().StoreReader(key, filename, size, reader, extra)
	if err != nil {
		return err
	}

	d.doReplicate(key)

	return nil
}

func (d *DHashChunkSystem) DeleteLocally(key []byte) error {

	if d.recover.isRecovering(key) {
		// 假如isRecovering为真，但minusCount失败，说明已经恢复完成，可以继续执行后面的函数
		if ok := d.recover.minusCount(key); ok {
			return nil
		}
	}

	info, err := d.local().GetInfo(key)
	if err != nil {
		return err
	}

	err = d.local().Delete(key)
	if err != nil {
		return err
	}

	return d.tryDeleteReplica(info)
}

func (d *DHashChunkSystem) Node() *node.Node {
	return d.ns.Self()
}

func (d *DHashChunkSystem) RecoverChunk(key []byte) (IDHashChunk, error) {
	if !d.Config().EnableReplica {
		return nil, hashchunk.ErrChunkNotFound
	}

	d.l.Info("Recover chunk", "key", hex.EncodeToString(key))

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
	d.l.Info("Found chunk", "chunk key", hex.EncodeToString(key))

	//先存到本地
	if err = func() error {
		//清理残余Chunk Info
		if err = d.local().DeleteInfo(key); err != nil {
			return err
		}

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
		return nil, util.WarpWithDetail(err)
	}

	hashChunk := &hashchunk.HashChunk{
		ReadSeekCloser: reader,
	}

	//chunk信息
	chunkInfo := replicaInfo.Custom
	replicaInfo.Custom = nil

	//还原包含chunk和replica信息的Info
	info := hashchunk.NewInfo(chunkInfo, hashchunk.NewExtraInfo("replica", replicaInfo))

	hashChunk.SetInfo(info)

	return &DHashChunk{
		HashChunk: hashChunk,
	}, nil
}

func (d *DHashChunkSystem) doReplicate(key []byte) {
	if d.config.EnableReplica {
		go func() {
			chunk, err := d.local().Get(key)
			if err != nil {
				d.l.Error("[Replica] get chunk failed", "error", err)
			}
			//后台执行副本存储
			if err = d.replicaService.PutRelicaTask(key,
				chunk,
				chunk.Info().ChunkInfo,
				func(res replica.TaskResult[*hashchunk.HashChunkInfo]) {
					if res.Err != nil {
						d.l.Error("Put replica task result", "error", res.Err)
					}

					//存储副本信息
					if err := d.local().UpdateInfo(key, func(info *hashchunk.Info) {
						info.ExtraInfo = hashchunk.NewExtraInfo("replica", res.ReplicaInfo)
					}); err != nil {
						d.l.Error("Update chunk info failed", "error", err)
					}

					chunk.Close()
				}); err != nil {
				d.l.Error("[Replica] Put replica task failed", "key", key, "error", err)
			}
		}()
	}
}

func (d *DHashChunkSystem) tryDeleteReplica(info *hashchunk.Info) error {
	// 如果chunk引用计数执行删除后为0，删除副本
	if info.ChunkInfo.ChunkCount-1 == 0 {

		replicaInfo, ok := info.ExtraInfo.TagInfo("replica")
		if !ok {
			return util.WarpWithDetail(replica.ErrReplicaInfoNotFound)
		}

		if replicaInfo == nil {
			return util.WarpWithDetail(replica.ErrReplicaInfoNotFound)
		}

		replicaInfo.Custom = info.ChunkInfo
		d.doDeleteReplica(replicaInfo)
	}
	return nil
}

func (d *DHashChunkSystem) doDeleteReplica(info *HashChunkReplicaInfo) {
	if d.config.EnableReplica {
		go func() {
			if err := d.replicaService.DeleteReplica(info); err != nil {
				d.l.Error("[Replica] Delete replica task failed", "error", err)
			}
		}()
	}
}

func (d *DHashChunkSystem) Config() *DHCSConfig {
	return d.config
}
