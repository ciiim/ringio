package tree

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/ciiim/cloudborad/storage/types"
)

const (
	SIZE_DIR = -1
)

type SubInfo struct {
	Index      int        `json:"index"`
	Name       string     `json:"name"`
	Size       types.Byte `json:"size"`
	IsDir      bool       `json:"file_type"`
	ModTime    time.Time  `json:"mod_time"`
	CreateTime time.Time
}

func DirEntryToSubInfo(baseDir string, de []fs.DirEntry) []*SubInfo {
	subList := make([]*SubInfo, len(de))
	for i, v := range de {
		info, _ := v.Info()
		fileSize := int64(0)
		var createTime time.Time
		if !v.IsDir() {
			func(entry fs.DirEntry) {
				file, err := os.Open(filepath.Join(baseDir, entry.Name()))
				if err != nil {
					return
				}
				defer file.Close()
				data := make([]byte, info.Size())
				if _, err := file.Read(data); err != nil {
					return
				}
				metadata := &Metadata{}
				if err := UnmarshalMetaData(data, metadata); err != nil {
					return
				}
				fileSize = metadata.Size
				createTime = metadata.CreateTime

			}(v)
		} else {
			fileSize = SIZE_DIR
		}
		subList[i] = &SubInfo{
			Index:      i,
			Name:       v.Name(),
			IsDir:      v.IsDir(),
			Size:       fileSize,
			ModTime:    info.ModTime(),
			CreateTime: createTime,
		}
	}
	return subList
}
