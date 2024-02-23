package tree

import (
	"encoding/json"
	"time"
)

const (
	META_FILE_SUFFIX = ".meta.json"
)

type Metadata struct {
	FileHash   []byte       `json:"file_meta_hash"`
	Filename   string       `json:"file_name"`
	Size       int64        `json:"file_size"`
	ModTime    time.Time    `json:"file_mod_time"`
	CreateTime time.Time    `json:"file_create_time"`
	Chunks     []*FileChunk `json:"chunks"`
}

type MetadataPath struct {
	Space string `json:"space"`
	Base  string `json:"base"`
	Name  string `json:"name"`
}

type FileChunk struct {

	// Begin from 0
	ChunkID int64 `json:"chunk_id"`

	Host string `json:"chunk_host"`
	Hash []byte `json:"chunk_hash"`
	Size int64  `json:"chunk_size"`

	Offset int64 `json:"chunk_offset"`
}

func NewMetaData(filename string, fileHash []byte, modTime time.Time, chunks []*FileChunk) *Metadata {
	m := &Metadata{
		Filename:   filename,
		FileHash:   fileHash,
		ModTime:    modTime,
		CreateTime: time.Now(),
		Chunks:     chunks,
	}
	var size int64
	for i := 0; i < len(m.Chunks); i++ {
		m.Chunks[i].ChunkID = int64(i)
		m.Chunks[i].Offset = size
		size += m.Chunks[i].Size
	}
	m.Size = size
	return m
}

func NewFileChunk(size int64, hash []byte) *FileChunk {
	return &FileChunk{
		Size: size,
		Hash: hash,
	}
}

func GetMetadataRealSize(metadata *Metadata) int64 {
	return metadata.Size
}

func UnmarshalMetaData(data []byte, metadata *Metadata) error {
	if err := json.Unmarshal(data, metadata); err != nil {
		return err
	}
	return nil
}

func MarshalMetaData(meta *Metadata) ([]byte, error) {
	data, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return data, nil
}
