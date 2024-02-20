package hashchunk

import (
	"io"
	"time"
)

type IHashChunkSystem interface {
	CreateChunk(key []byte, name string, extra *ExtraInfo) (io.WriteCloser, error)
	StoreBytes(key []byte, name string, value []byte, extra *ExtraInfo) error
	StoreReader(key []byte, name string, v io.Reader, extra *ExtraInfo) error
	Get(key []byte) (*HashChunk, error)
	Delete(key []byte) error
	Cap() int64
	Occupied(unit ...string) float64
	Config() *Config
}

type IHashChunk interface {
	io.ReadCloser
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
