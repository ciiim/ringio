package hashchunk

import (
	"io"
	"time"
)

type IHashChunkSystem interface {
	CreateChunk(key []byte, name string, size int64, extra *ExtraInfo) (io.WriteCloser, error)
	StoreBytes(key []byte, name string, size int64, value []byte, extra *ExtraInfo) error
	StoreReader(key []byte, name string, size int64, v io.Reader, extra *ExtraInfo) error

	Get(key []byte) (*HashChunk, error)
	Delete(key []byte) error

	GetInfo(key []byte) (*Info, error)
	UpdateInfo(key []byte, updateFn func(info *Info)) error
	DeleteInfo(key []byte) error

	Cap() int64
	Occupied(unit ...string) float64
	Config() *Config
}

type IHashChunk interface {
	io.ReadSeekCloser
	Info() *Info
}

type IHashChunkInfo interface {
	Name() string
	Count() int64
	Path() string //base path
	Hash() []byte //chunk's hash
	Size() int64
	ModTime() time.Time
	CreateTime() time.Time
}
