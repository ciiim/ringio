package hashchunk

import (
	"time"
)

const (
	RemoteChunkCount = -1
)

type HashChunkInfo struct {

	// 块的引用计数
	// -1 表示为远程块文件
	ChunkCount int64 `json:"count"`

	ChunkName       string    `json:"chunk_name"`
	ChunkHash       []byte    `json:"hash"`
	ChunkPath       string    `json:"path"`
	ChunkSize       int64     `json:"size"`
	ChunkModTime    time.Time `json:"mod_time"`
	ChunkCreateTime time.Time `json:"create_time"`
}

func NewChunkInfo(chunkName string, hashSum []byte, size int64) *HashChunkInfo {
	now := time.Now()
	hcstat := &HashChunkInfo{
		ChunkName:       chunkName,
		ChunkHash:       hashSum,
		ChunkSize:       size,
		ChunkModTime:    now,
		ChunkCreateTime: now,
	}
	hcstat.ChunkCount = 1
	return hcstat
}

func (hcstat *HashChunkInfo) SetPath(path string) *HashChunkInfo {
	hcstat.ChunkPath = path
	return hcstat
}

func (hcstat *HashChunkInfo) Name() string {
	return hcstat.ChunkName
}

func (hcstat *HashChunkInfo) Path() string {
	return hcstat.ChunkPath
}

func (hcstat *HashChunkInfo) Hash() []byte {
	return hcstat.ChunkHash
}

func (hcstat *HashChunkInfo) Size() int64 {
	return int64(hcstat.ChunkSize)
}

func (hcstat *HashChunkInfo) ModTime() time.Time {
	return hcstat.ChunkModTime
}

func (hcstat *HashChunkInfo) CreateTime() time.Time {
	return hcstat.ChunkCreateTime
}
