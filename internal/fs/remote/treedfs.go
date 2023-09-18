package remote

import (
	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/internal/fs/peers"
)

const (
	NEW_SPACE = "__NEW__SPACE__"
)

// Distributed Tree File System.
// Implement FileSystem interface
type TreeDFileSystem struct {
	*fs.TreeFileSystem
	remote *rpcTreeClient
	self   peers.Peer
}

var _ TreeDFileSystemI = (*TreeDFileSystem)(nil)

func NewTreeDFileSystem(rootPath string) *TreeDFileSystem {
	TreeDFileSystem := &TreeDFileSystem{
		remote:         newRPCTreeClient(),
		TreeFileSystem: fs.NewTreeFileSystem(rootPath),
	}
	return TreeDFileSystem

}

func (dt *TreeDFileSystem) SetPeerService(ps peers.Peer) {
	dt.self = ps
}

func (dt *TreeDFileSystem) pickPeer(key string) peers.PeerInfo {
	return dt.self.Pick(key)
}

func (dt *TreeDFileSystem) NewSpace(space string, cap fs.Byte) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.NewSpace(space, cap)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.newSpace(ctx, pi, space, cap)
}

func (dt *TreeDFileSystem) DeleteSpace(space string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.DeleteSpace(space)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.deleteSpace(ctx, pi, space)
}

func (dt *TreeDFileSystem) MakeDir(space, base, name string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.MakeDir(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.makeDir(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) RenameDir(space, base, name, newName string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		dt.TreeFileSystem.RenameDir(space, base, name, newName)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.renameDir(ctx, pi, space, base, name, newName)
}

func (dt *TreeDFileSystem) DeleteDir(space, base, name string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.DeleteDir(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.deleteDir(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) GetDirSub(space, base, name string) ([]fs.SubInfo, error) {
	pi := dt.pickPeer(space)
	if pi == nil {
		return nil, fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.GetDirSub(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.getDirSub(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) GetMetadata(space, base, name string) ([]byte, error) {
	pi := dt.pickPeer(space)
	if pi == nil {
		return nil, fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.GetMetadata(space, base, name)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.getMetadata(ctx, pi, space, base, name)
}

func (dt *TreeDFileSystem) PutMetadata(space, base, name, hash string, data []byte) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.PutMetadata(space, base, name, hash, data)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.putMetadata(ctx, pi, space, base, name, data)
}
func (dt *TreeDFileSystem) DeleteMetadata(space, base, name, hash string) error {
	pi := dt.pickPeer(space)
	if pi == nil {
		return fs.ErrSpaceNotFound
	}
	if dt.self.Info().Equal(pi) {
		return dt.TreeFileSystem.DeleteMetadata(space, base, name, hash)
	}
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return dt.remote.deleteMetadata(ctx, pi, space, base, name, hash)
}

func (dt *TreeDFileSystem) Peer() peers.Peer {
	return dt.self
}
