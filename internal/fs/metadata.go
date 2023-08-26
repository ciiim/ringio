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

type Fileblock struct {

	// Begin from 0
	BlockID int64 `json:"block_id"`

	Host string `json:"block_host"`
	Hash string `json:"block_hash"`
	Size int64  `json:"block_size"`
}

func newMetaData(filename string, hash string, modTime time.Time, blocks []Fileblock) Metadata {
	m := Metadata{
		Filename: filename,
		Hash:     hash,
		ModTime:  modTime,
		Blocks:   blocks,
	}
	var size int64
	blockID := int64(0)
	for i := 0; i < len(blocks); i++ {
		blocks[i].BlockID = blockID
		blockID++
		size += blocks[i].Size
	}
	m.Size = size
	return m
}

func newFileBlock(host string, size int64, hash string) Fileblock {
	return Fileblock{
		Host: host,
		Size: size,
		Hash: hash,
	}
}

func unmarshalMetaData(data []byte, metadata *Metadata) error {
	if err := json.Unmarshal(data, metadata); err != nil {
		return err
	}
	return nil
}

func marshalMetaData(meta *Metadata) ([]byte, error) {
	data, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return data, nil
}
