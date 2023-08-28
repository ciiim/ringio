package fs

import (
	"log"

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

func (dt *TreeDFileSystem) Serve() {
	log.Println("[TreeDFileSystem] Serve on", dt.self.PAddr())
	newRPCTDFSServer(dt).serve(RPC_TDFS_PORT)
}
