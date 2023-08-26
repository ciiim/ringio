package fs

import (
	"log"
	"sync"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

const (
	NEW_SPACE = "__NEW__SPACE__"
)

// Distributed Tree File System.
// Implement FileSystem interface
type TreeDFileSystem struct {
	*treeFileSystem
	remote *rpcTreeClient
	self   DPeer
}

type DTreeFile struct {
	data []byte
	info TreeDFileInfo
}
type TreeDFileInfo struct {
	TreeFileInfo
	DPeerInfo
}

var _ TreeDFileSystemI = (*TreeDFileSystem)(nil)

func NewTreeDFileSystem(self DPeer, rootPath string) *TreeDFileSystem {
	TreeDFileSystem := &TreeDFileSystem{
		self:           self,
		remote:         newRPCTreeClient(),
		treeFileSystem: newTreeFileSystem(rootPath),
	}
	return TreeDFileSystem

}

func (dt *TreeDFileSystem) pickPeer(key string) peers.PeerInfo {
	return dt.self.Pick(key)
}

func (dt *TreeDFileSystem) NewSpace(space string, cap Byte) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.NewSpace(space, cap)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.newSpace(ctx, pi, space, cap)
}

func (dt *TreeDFileSystem) DeleteSpace(space string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.DeleteSpace(space)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.deleteSpace(ctx, pi, space)
}

func (dt *TreeDFileSystem) MakeDir(space, base, name string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.MakeDir(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.makeDir(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) RenameDir(space, base, name, newName string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		dt.treeFileSystem.RenameDir(space, base, name, newName)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.renameDir(ctx, pi, space, base, name, newName)
}

func (dt *TreeDFileSystem) DeleteDir(space, base, name string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.DeleteDir(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.deleteDir(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) GetDirSub(space, base, name string) ([]SubInfo, error) {
	pi := dt.pickPeer(space)
	if pi == nil {
		return nil, ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.GetDirSub(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.getDirSub(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) GetMetadata(space, base, name string) ([]byte, error) {
	pi := dt.pickPeer(space)
	if pi == nil {
		return nil, ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.GetMetadata(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.getMetadata(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) HasSameMetadata(hash string) (MetadataPath, bool) {
	path, has := dt.treeFileSystem.HasSameMetadata(hash)
	if has {
		return path, has
	}
	//向所有节点查询
	list := dt.self.PList()
	if len(list) == 1 {
		return MetadataPath{}, false
	}
	respChan := make(chan bool, len(list)-1)
	wgDoneChan := make(chan struct{})
	var resPath MetadataPath
	var wg sync.WaitGroup
	wg.Add(len(list) - 1)
	go func() {
		wg.Wait()
		wgDoneChan <- struct{}{}
	}()
	for _, pi := range list {
		go func(pi_ peers.PeerInfo) {
			if dt.self.info.Equal(pi_) {
				return
			}
			ctx, cancel := ctxWithTimeout()
			defer cancel()
			path, has = dt.remote.hasSameMetadata(ctx, pi_, hash)
			resPath = path
			respChan <- has
			wg.Done()
		}(pi)
	}
	for {
		select {
		case has := <-respChan:
			if has {
				close(respChan)
				close(wgDoneChan)
				return resPath, true
			}
		case <-wgDoneChan:
			close(respChan)
			close(wgDoneChan)
			return MetadataPath{}, false
		}
	}
}

func (dt *TreeDFileSystem) HasSameMetadataLocal(hash string) (MetadataPath, bool) {
	return dt.treeFileSystem.HasSameMetadata(hash)
}

func (dt *TreeDFileSystem) PutMetadata(space, base, name, hash string, data []byte) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.PutMetadata(space, base, name, hash, data)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.putMetadata(ctx, pi, space, base, name, data)
}
func (dt *TreeDFileSystem) DeleteMetadata(space, base, name, hash string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return ErrSpaceNotFound
	}
	if dt.self.info.Equal(pi) {
		return dt.treeFileSystem.DeleteMetadata(space, base, name, hash)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.deleteMetadata(ctx, pi, space, base, name, hash)
}

func (dt *TreeDFileSystem) Peer() peers.Peer {
	return dt.self
}

func (df DTreeFile) Metadata() []byte {
	return df.data
}

func (df DTreeFile) Stat() TreeDFileInfo {
	return df.info
}

func (dfi TreeDFileInfo) PeerInfo() peers.PeerInfo {
	return dfi.DPeerInfo
}

func (dt *TreeDFileSystem) Serve() {
	log.Println("[TreeDFileSystem] Serve on", dt.self.PAddr())
	newRPCTDFSServer(dt).Serve(RPC_TDFS_PORT)
}
