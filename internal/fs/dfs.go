package fs

import (
	"errors"
	"fmt"
	"log"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

type DistributeFile struct {
	data []byte
	info DistributeFileInfo
}

type DistributeFileInfo struct {
	BasicFileInfo
	DPeerInfo
}

// distribute file system
type DFS struct {
	*basicFileSystem
	self peers.Peer
}

var _ DistributeFileSystem = (*DFS)(nil)
var _ FileInfo = (*DistributeFileInfo)(nil)

func (d *DFS) AddPeer(pis ...peers.PeerInfo) error {
	d.self.PAdd(pis...)
	return nil
}

func (d *DFS) PickPeer(key string) peers.PeerInfo {
	return d.self.Pick(key)
}

func NewDFS(self peers.Peer, rootPath string, capacity int64, calcStorePathFn CalcStoreFilePathFnType) *DFS {
	d := &DFS{
		basicFileSystem: newBasicFileSystem(rootPath, capacity, calcStorePathFn),

		self: self.(DPeer),
	}
	return d
}

func (d *DFS) Get(key string) (File, error) {
	dlog.debug("[DFS]", "Get by key:", key)
	pi := d.PickPeer(key)
	if pi == nil {
		return nil, peers.ErrPeerNotFound
	}
	// get from local
	if pi.Equal(d.self.Info()) {
		log.Println("[DFS]Get from local.")
		df, err := d.getLocally(key)
		if errors.Is(err, ErrFileNotFound) {
			return d.recoverFile(key)
		} else {
			return df, err
		}
	}

	// no peer
	if pi.Equal(DPeerInfo{}) {
		return DistributeFile{}, fmt.Errorf("no peer for key %s", key)
	}

	// get from remote
	log.Println("[DFS]Get from remote.")
	resp := d.self.Get(pi, key)
	return DistributeFile{
		data: resp.Data,
		info: resp.Info.(DistributeFileInfo),
	}, resp.Err
}

func (d *DFS) Store(key string, filename string, value []byte) error {
	dlog.debug("[DFS]", "Store by key:", key, " filename:", filename, " value len:", len(value))
	pi := d.PickPeer(key)
	if pi == nil {
		return peers.ErrPeerNotFound
	}
	// store locally
	if pi.Equal(d.self.Info()) {
		log.Println("[DFS]Store locally.")
		return d.storeLocally(key, filename, value)
	}

	// no peer
	if pi.Equal(DPeerInfo{}) {
		return fmt.Errorf("no peer for key %s", key)
	}

	// store remotely
	log.Println("[DFS]Put to remote")
	return d.self.Put(pi, key, filename, value).Err
}

func (d *DFS) Delete(key string) error {
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
	return d.self.Delete(pi, key).Err
}

func (d *DFS) getLocally(key string) (DistributeFile, error) {
	file, err := d.basicFileSystem.Get(key)
	fi := file.Stat()
	return DistributeFile{
			data: file.Data(),
			info: DistributeFileInfo{BasicFileInfo: fi.(BasicFileInfo), DPeerInfo: d.self.Info().(DPeerInfo)},
		},
		err
}

func (d *DFS) storeLocally(key string, filename string, value []byte) error {
	return d.basicFileSystem.Store(key, filename, value)
}

func (d *DFS) deleteLocally(key string) error {
	return d.basicFileSystem.Delete(key)
}

func (d *DFS) Peer() peers.Peer {
	return d.self
}

/*
will happen when new peer join the cluster
*/
func (d *DFS) recoverFile(key string) (File, error) {
	nextInfo := d.Peer().PNext(key)
	if nextInfo == nil {
		return nil, peers.ErrPeerNotFound
	}
	if nextInfo.Equal(d.self.Info()) {
		return nil, ErrFileNotFound
	}
	// Get file info from next peer
	resp := d.self.Get(nextInfo, key)
	if resp.Err == nil {
		// delete the wrong local file
		return DistributeFile{
			data: resp.Data,
			info: resp.Info.(DistributeFileInfo),
		}, nil
	}
	return nil, resp.Err
}

func (df DistributeFile) Data() []byte {
	return df.data
}

func (df DistributeFile) Stat() FileInfo {
	return df.info
}

func (dfi DistributeFileInfo) PeerInfo() peers.PeerInfo {
	return dfi.DPeerInfo
}

func (dfi DistributeFileInfo) SubDir() []SubInfo {
	return dfi.BasicFileInfo.SubDir()
}

func (d *DFS) Serve() {
	log.Println("[DFS] Serve on", d.self.PAddr())
	newRpcServer(d).run(FILE_STORE_PORT)
}
