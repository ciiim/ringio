package fs

import (
	"errors"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

var (
	ErrFull            = errors.New("file system is full")
	ErrFileNotFound    = errors.New("file not found")
	ErrFileExist       = errors.New("file or dir already exist")
	ErrFileInvalidName = errors.New("invalid file name")
	ErrNotDir          = errors.New("not a directory")
	ErrIsDir           = errors.New("is a directory")
	ErrInternal        = errors.New("internal error")
)

type HashFileSystemI interface {
	Store(key, name string, value []byte) error
	Get(key string) (HashFileI, error)
	Delete(key string) error
	Cap() int64
	Occupied(unit ...string) float64
	Close() error
}

type HashDFileSystemI interface {
	HashFileSystemI
	Peer() peers.Peer
}

type TreeFileSystemI interface {
	NewSpace(space string, cap Byte) error
	GetSpace(space string) *Space
	DeleteSpace(space string) error
	Close() error
}

type TreeDFileSystemI interface {
	TreeFileSystemI

	MakeDir(space, base, name string) error
	RenameDir(space, base, name, newName string) error
	DeleteDir(space, base, name string) error
	GetDirSub(space, base, name string) ([]SubInfo, error)

	GetMetadata(space, base, name string) ([]byte, error)
	PutMetadata(space, base, name, hash string, data []byte) error
	DeleteMetadata(space, base, name, hash string) error

	Peer() peers.Peer
}

type DefaultHashDFileSystem struct{}

var _ HashDFileSystemI = (*DefaultHashDFileSystem)(nil)

func (DefaultHashDFileSystem) Store(key, name string, value []byte) error {
	return nil
}

func (DefaultHashDFileSystem) Get(key string) (HashFileI, error) {
	return nil, nil
}

func (DefaultHashDFileSystem) Delete(key string) error {
	return nil
}

func (DefaultHashDFileSystem) Cap() int64 {
	return 0
}

func (DefaultHashDFileSystem) Occupied(unit ...string) float64 {
	return 0
}

func (DefaultHashDFileSystem) Close() error {
	return nil
}

func (DefaultHashDFileSystem) Peer() peers.Peer {
	return nil
}

type DefaultTreeDFileSystem struct{}

var _ TreeDFileSystemI = (*DefaultTreeDFileSystem)(nil)

func (DefaultTreeDFileSystem) NewSpace(space string, cap Byte) error {
	return nil
}

func (DefaultTreeDFileSystem) GetSpace(space string) *Space {
	return nil
}

func (DefaultTreeDFileSystem) DeleteSpace(space string) error {
	return nil
}

func (DefaultTreeDFileSystem) Close() error {
	return nil
}

func (DefaultTreeDFileSystem) MakeDir(space, base, name string) error {
	return nil
}

func (DefaultTreeDFileSystem) RenameDir(space, base, name, newName string) error {
	return nil
}

func (DefaultTreeDFileSystem) DeleteDir(space, base, name string) error {
	return nil
}

func (DefaultTreeDFileSystem) GetDirSub(space, base, name string) ([]SubInfo, error) {
	return nil, nil
}

func (DefaultTreeDFileSystem) GetMetadata(space, base, name string) ([]byte, error) {
	return nil, nil
}

func (DefaultTreeDFileSystem) PutMetadata(space, base, name, hash string, data []byte) error {
	return nil
}

func (DefaultTreeDFileSystem) DeleteMetadata(space, base, name, hash string) error {
	return nil
}

func (DefaultTreeDFileSystem) Peer() peers.Peer {
	return nil
}
