package ringio

import (
	"io"

	"github.com/ciiim/cloudborad/storage/hashchunk"
	"github.com/ciiim/cloudborad/storage/tree"
	"github.com/ciiim/cloudborad/storage/types"
)

type IDHashChunk interface {
	hashchunk.IHashChunk
}

type IDHashChunkSystem interface {

	// 本地入口方法
	//
	// 可绕过副本服务
	local() hashchunk.IHashChunkSystem

	Config() *DHCSConfig

	Get(key []byte) (IDHashChunk, error)

	GetLocally(key []byte) (IDHashChunk, error)

	Delete(key []byte) error

	DeleteLocally(key []byte) error

	CreateChunkLocally(key []byte, name string, size int64, extra *hashchunk.ExtraInfo) (io.WriteCloser, error)

	StoreReader(key []byte, name string, size int64, reader io.Reader, extra *hashchunk.ExtraInfo) error

	StoreBytes(key []byte, name string, size int64, v []byte, extra *hashchunk.ExtraInfo) error

	// 从其它节点查找chunk
	FindChunk(chunkKey []byte) (IDHashChunk, error)
}

type ITreeDFileSystem interface {

	// 本地入口方法
	Local() tree.ITreeFileSystem

	// 新建Space
	NewSpace(space string, cap types.Byte) error

	// 删除Space
	DeleteSpace(space string) error

	// Space 操作
	MakeDir(space, base, dir string) error
	RenameDir(space, base, dir, newDirName string) error
	DeleteDir(space, base, dir string) error
	GetDirSub(space, base, dir string) ([]*tree.SubInfo, error)

	GetMetadata(space, base, name string) ([]byte, error)
	PutMetadata(space, base, name string, hash []byte, data []byte) error
	DeleteMetadata(space, base, name string) error
}
