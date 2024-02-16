package hashchunk

import (
	"io"
	"time"
)

type IHashChunkSystem interface {
	CreateChunk(key []byte, name string) (io.WriteCloser, error)
	StoreBytes(key []byte, name string, value []byte) error
	StoreReader(key []byte, name string, v io.Reader) error
	Get(key []byte) (*HashChunk, error)
	Delete(key []byte) error
	Cap() int64
	Occupied(unit ...string) float64
	Config() *Config
}

type IHashChunk interface {
	io.ReadCloser
	Info() IHashChunkInfo
}

type IHashChunkInfo interface {
	Name() string
	Path() string //base path
	Hash() []byte //chunk's hash
	Size() int64
	ModTime() time.Time
	CreateTime() time.Time
}
