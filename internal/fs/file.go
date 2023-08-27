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
