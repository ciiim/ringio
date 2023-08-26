package fs

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

type HashDFile struct {
	data []byte
	info HashDFileInfo
}

type HashDFileInfo struct {
	HashFileInfo
	DPeerInfo
}

// distribute file system
type HashDFileSystem struct {
	*hashFileSystem
	remote *rpcHashClient
	self   peers.Peer
}

var _ HashDFileSystemI = (*HashDFileSystem)(nil)
var _ HashFileInfoI = (*HashDFileInfo)(nil)

func (d *HashDFileSystem) AddPeer(pis ...peers.PeerInfo) error {
	d.self.PAdd(pis...)
	return nil
}

func (d *HashDFileSystem) PickPeer(key string) peers.PeerInfo {
	return d.self.Pick(key)
}

func NewDFS(self peers.Peer, rootPath string, capacity int64, calcStorePathFn CalcStoreFilePathFnType) *HashDFileSystem {
	d := &HashDFileSystem{
		hashFileSystem: newHashFileSystem(rootPath, capacity, calcStorePathFn),
		remote:         newRPCHashClient(),
		self:           self.(DPeer),
	}
	return d
}

func (d *HashDFileSystem) Get(key string) (HashFileI, error) {
	dlog.debug("[HashDFileSystem]", "Get by key '%s'", key)
	pi := d.PickPeer(key)
	if pi == nil {
		return nil, peers.ErrPeerNotFound
	}
	// get from local
	if pi.Equal(d.self.Info()) {
		log.Println("[HashDFileSystem]Get from local.")
		df, err := d.getLocally(key)
		if errors.Is(err, ErrFileNotFound) {
			return d.recoverFile(key)
		} else {
			return df, err
		}
	}

	// no peer
	if pi.Equal(DPeerInfo{}) {
		return HashDFile{}, fmt.Errorf("no peer for key %s", key)
	}

	// get from remote
	log.Println("[HashDFileSystem]Get from remote.")
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()

	resp, err := d.remote.get(ctx, pi, key)
	return HashDFile{
		data: resp.Data(),
		info: resp.Stat().(HashDFileInfo),
	}, err
}

func (d *HashDFileSystem) Store(key string, filename string, value []byte) error {
	dlog.debug("[HashDFileSystem]", "Store by key '%s', name '%s'", key, filename)
	pi := d.PickPeer(key)
	if pi == nil {
		return peers.ErrPeerNotFound
	}
	// store locally
	if pi.Equal(d.self.Info()) {
		log.Println("[HashDFileSystem]Store locally.")
		return d.storeLocally(key, filename, value)
	}

	// no peer
	if pi.Equal(DPeerInfo{}) {
		return fmt.Errorf("no peer for key %s", key)
	}

	// store remotely
	log.Println("[HashDFileSystem]Put to remote")
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.put(ctx, pi, key, filename, value)
}

func (d *HashDFileSystem) Delete(key string) error {
	pi := d.PickPeer(key)
	if pi == nil {
		return peers.ErrPeerNotFound
	}
	// delete locally
	if pi.Equal(d.self.Info()) {
		return d.deleteLocally(key)
	}

	// no peer
	if pi.Equal(DPeerInfo{}) {
		return fmt.Errorf("no peer for key %s", key)
	}

	// delete remotely
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return d.remote.delete(ctx, pi, key)
}

func (d *HashDFileSystem) getLocally(key string) (HashDFile, error) {
	file, err := d.hashFileSystem.Get(key)
	fi := file.Stat()
	return HashDFile{
			data: file.Data(),
			info: HashDFileInfo{HashFileInfo: fi.(HashFileInfo), DPeerInfo: d.self.Info().(DPeerInfo)},
		},
		err
}

func (d *HashDFileSystem) storeLocally(key string, filename string, value []byte) error {
	return d.hashFileSystem.Store(key, filename, value)
}

func (d *HashDFileSystem) deleteLocally(key string) error {
	return d.hashFileSystem.Delete(key)
}

func (d *HashDFileSystem) Peer() peers.Peer {
	return d.self
}

/*
will happen when new peer join the cluster
*/
func (d *HashDFileSystem) recoverFile(key string) (HashDFile, error) {
	nextInfo := d.Peer().PNext(key)
	if nextInfo == nil {
		return HashDFile{}, peers.ErrPeerNotFound
	}
	if nextInfo.Equal(d.self.Info()) {
		return HashDFile{}, ErrFileNotFound
	}
	// Get file info from next peer
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	resp, err := d.remote.get(ctx, nextInfo, key)
	if err == nil {
		// delete the wrong local file
		return HashDFile{
			data: resp.Data(),
			info: resp.Stat().(HashDFileInfo),
		}, nil
	}
	return HashDFile{}, err
}

func (df HashDFile) Data() []byte {
	return df.data
}

func (df HashDFile) Stat() HashFileInfoI {
	return df.info
}

func (dfi HashDFileInfo) PeerInfo() peers.PeerInfo {
	return dfi.DPeerInfo
}

func (d *HashDFileSystem) Serve() {
	log.Println("[HashDFileSystem] Serve on", d.self.PAddr())
	newRPCHDFSServer(d).run(RPC_HDFS_PORT)
}
