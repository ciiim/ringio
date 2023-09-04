package fs

import (
	"io/fs"
	"time"
)

type SubInfo struct {
	Index   int       `json:"index"`
	Name    string    `json:"dir_name"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
}

func DirEntryToSubInfo(de []fs.DirEntry) []SubInfo {
	var subList []SubInfo
	for i, v := range de {
		info, _ := v.Info()
		subList = append(subList, SubInfo{
			Index:   i,
			Name:    v.Name(),
			IsDir:   v.IsDir(),
			ModTime: info.ModTime(),
		})
	}
	return subList
}
