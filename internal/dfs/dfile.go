package dfs

import (
	"time"

	"github.com/ciiim/cloudborad/internal/dfs/peers"
)

type HashDFileI interface {
	Data() []byte
	Stat() HashDFileInfoI
}

type HashDFileInfoI interface {
	Name() string
	Path() string //base path
	Hash() string //file's hash
	Size() int64
	ModTime() time.Time

	PeerInfo() peers.PeerInfo
}
