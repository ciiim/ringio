package fs

import (
	"time"
)

type Byte = int64

const (
	MB = Byte(8 << 17)
	GB = Byte(8 << 27)
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
}
