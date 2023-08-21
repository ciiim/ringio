package fs

import (
	"encoding/json"
	"time"
)

const (
	META_FILE_SUFFIX = ".meta"
)

type Metadata struct {
	Filename string      `json:"filename"`
	Hash     string      `json:"hash"`
	Size     int64       `json:"size"`
	ModTime  time.Time   `json:"mod_time"`
	Blocks   []Fileblock `json:"blocks"`
}

type Fileblock struct {

	// Begin from 0
	BlockID int64 `json:"blockid"`

	/*
	 format: <hostip>@<block_path>

	 block_path : <block_dir>/<block_name>
	*/
	FullPath string `json:"fullpath"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
}

func newMetaData(filename string, hash string, size int64, modTime time.Time, blocks []Fileblock) Metadata {
	return Metadata{
		Filename: filename,
		Hash:     hash,
		Size:     size,
		ModTime:  modTime,
		Blocks:   blocks,
	}
}

func newFileBlock(fullPath string, size int64, hash string) Fileblock {
	return Fileblock{
		FullPath: fullPath,
		Size:     size,
		Hash:     hash,
	}
}

func readMetaDataByBytes(data []byte, metadata *Metadata) error {
	if err := json.Unmarshal(data, metadata); err != nil {
		return err
	}
	return nil
}

func marshalMetaData(meta Metadata) []byte {
	data, err := json.Marshal(meta)
	if err != nil {
		return nil
	}
	return data
}
