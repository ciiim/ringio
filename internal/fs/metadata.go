package fs

import (
	"encoding/json"
	"time"
)

const (
	META_FILE_SUFFIX = ".meta"
)

type Metadata struct {
	Hash     string      `json:"file_meta_hash"`
	Filename string      `json:"file_name"`
	Size     int64       `json:"file_size"`
	ModTime  time.Time   `json:"file_mod_time"`
	Blocks   []Fileblock `json:"block_list"`
}

type MetadataPath struct {
	Space string `json:"space"`
	Base  string `json:"base"`
	Name  string `json:"name"`
}

type Fileblock struct {

	// Begin from 0
	BlockID int64 `json:"block_id"`

	Host string `json:"block_host"`
	Hash string `json:"block_hash"`
	Size int64  `json:"block_size"`

	Offset int64 `json:"block_offset"`
}

func NewMetaData(filename string, hash string, modTime time.Time, blocks []Fileblock) Metadata {
	m := Metadata{
		Filename: filename,
		Hash:     hash,
		ModTime:  modTime,
		Blocks:   blocks,
	}
	var size int64
	for i := 0; i < len(m.Blocks); i++ {
		m.Blocks[i].BlockID = int64(i)
		m.Blocks[i].Offset = size
		size += m.Blocks[i].Size
	}
	m.Size = size
	return m
}

func NewFileBlock(host string, size int64, hash string) Fileblock {
	return Fileblock{
		Host: host,
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
