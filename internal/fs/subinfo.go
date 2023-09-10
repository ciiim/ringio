package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type FILE_TYPE string

var (
	TYPE_DIR    FILE_TYPE = "dir"
	TYPE_NORMAL FILE_TYPE = "file"
)

type SubInfo struct {
	Index   int       `json:"index"`
	Name    string    `json:"name"`
	Size    Byte      `json:"size"`
	Type    FILE_TYPE `json:"file_type"`
	ModTime time.Time `json:"mod_time"`
}

func GetFileType(isDir bool) FILE_TYPE {
	if isDir {
		return TYPE_DIR
	}
	return TYPE_NORMAL
}

func DirEntryToSubInfo(baseDir string, de []fs.DirEntry) []SubInfo {
	var subList []SubInfo
	for i, v := range de {
		info, _ := v.Info()
		fileSize := int64(0)
		if !v.IsDir() {
			file, _ := os.Open(filepath.Join(baseDir, v.Name()))
			data := make([]byte, info.Size())
			file.Read(data)
			metadata := &Metadata{}
			UnmarshalMetaData(data, metadata)
			fileSize = GetMetadataRealSize(metadata)
			file.Close()
		} else {
			fileSize = -1
		}
		subList = append(subList, SubInfo{
			Index:   i,
			Name:    v.Name(),
			Type:    GetFileType(v.IsDir()),
			Size:    fileSize,
			ModTime: info.ModTime(),
		})
	}
	return subList
}
