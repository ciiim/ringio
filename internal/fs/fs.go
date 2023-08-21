package fs

import (
	"errors"
	"io/fs"
	"time"

	"github.com/ciiim/cloudborad/internal/fs/peers"

	pb "github.com/ciiim/cloudborad/internal/fs/fspb"
)

var (
	ErrFull            = errors.New("file system is full")
	ErrFileNotFound    = errors.New("file not found")
	ErrFileExist       = errors.New("file or dir already exist")
	ErrFileInvalidName = errors.New("invalid file name")
	ErrNotDir          = errors.New("not a directory")
	ErrInternal        = errors.New("internal error")
)

type FileSystem interface {
	Store(key, name string, value []byte) error

	Get(key string) (File, error)

	Delete(key string) error

	Set(opt any) error

	Close() error
}

type DistributeFileSystem interface {
	FileSystem
	Serve()
	Peer() peers.Peer
}

type File interface {
	Data() []byte
	Stat() FileInfo
}

type FileInfo interface {
	Name() string
	Path() string //base path
	Hash() string //file's hash
	Size() int64
	ModTime() time.Time
	IsDir() bool

	PeerInfo() peers.PeerInfo

	SubDir() []fs.DirEntry
}

type Byte = int64

func pBFileInfoToBasicFileInfo(pb *pb.FileInfo) BasicFileInfo {
	if pb == nil {
		return BasicFileInfo{}
	}
	return BasicFileInfo{
		Path_:    pb.BasePath,
		FileName: pb.FileName,
		Hash_:    pb.Hash,
		Size_:    pb.Size,
		Dir_:     pb.IsDir,
	}
}

func pbFileInfoToTreeFileInfo(pb *pb.FileInfo) TreeFileInfo {
	if pb == nil {
		return TreeFileInfo{}
	}
	var subDir []fs.DirEntry
	if pb.IsDir {
		subDir = make([]fs.DirEntry, 0, len(pb.DirInfo))
		for _, v := range pb.DirInfo {
			subDir = append(subDir, Dir{v.GetDirName()})
		}
	}
	return TreeFileInfo{
		BasicFileInfo: pBFileInfoToBasicFileInfo(pb),
		subDir:        subDir,
	}
}
