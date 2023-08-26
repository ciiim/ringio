package fs

import (
	"time"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

type Byte = int64

const (
	MB = Byte(1024 * 1024)
	GB = Byte(1024 * 1024 * 1024)
)

type HashFileI interface {
	Data() []byte
	Stat() HashFileInfoI
}

type HashFileInfoI interface {
	Name() string
	Path() string //base path
	Hash() string //file's hash
	Size() int64
	ModTime() time.Time

	PeerInfo() peers.PeerInfo
}

type TreeFileI interface {
	Metadata() []byte
	Stat() TreeFileInfoI
}

type TreeFileInfoI interface {
	Name() string
	Path() string //base path
	Size() int64
	ModTime() time.Time
	IsDir() bool
	PeerInfo() peers.PeerInfo
	Sub() []SubInfo
}
