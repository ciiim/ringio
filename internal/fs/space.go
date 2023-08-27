package fs

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrSpaceExist    = fmt.Errorf("space exist")
	ErrSpaceFull     = fmt.Errorf("space full")
	ErrSpaceNotFound = fmt.Errorf("space not found")
	ErrSpaceInternal = fmt.Errorf("space internal error")
)

type Space struct {
	root string

	// /treeFS.rootPath/spaceKey
	spaceKey string
	base     string
	capacity Byte
	occupy   Byte
}

func (s *Space) storeMetaData(base, fileName string, metadata []byte) (err error) {

	return s.i_storeMetadata(base, fileName, metadata, os.O_CREATE|os.O_WRONLY)

}

func (s *Space) getMetadata(base, fileName string) ([]byte, error) {
	return s.i_getMetadata(base, fileName)
}

func (s *Space) deleteMetaData(base, fileName string) error {

	//TODO: 防止删除fullpath的上级目录
	size, err := s.GetSize(base, fileName)
	if err != nil {
		return err
	}
	if s.occupy < size {
		return ErrInternal
	}
	s.occupy -= int64(size)

	return os.Remove(s.getFullPath(base, fileName))
}

func (s *Space) makeDir(base, fileName string) error {

	//TODO: 防止创建到dir的上级目录内

	return os.Mkdir(s.getFullPath(base, fileName), 0755)
}

func (s *Space) renameDir(base, fileName, newName string) error {
	return os.Rename(s.getFullPath(base, fileName), s.getFullPath(base, newName))
}

func (s *Space) deleteDir(base, fileName string) error {
	return os.RemoveAll(s.getFullPath(base, fileName))
}

func (s *Space) getDir(base, fileName string) ([]fs.DirEntry, error) {

	//TODO: 防止访问dir的上级目录
	return os.ReadDir(s.getFullPath(base, fileName))
}

func (s *Space) i_storeMetadata(base, fileName string, metadata []byte, flag int) error {
	file, err := os.OpenFile(s.getFullPath(base, fileName), flag, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	info, _ := file.Stat()
	oldSize := info.Size()
	_, err = file.Write(metadata)
	if err != nil {
		return err
	}
	newSize := int64(len(metadata))
	s.occupy += Byte(newSize - oldSize)

	return nil
}

func (s *Space) i_getMetadata(base, fileName string) ([]byte, error) {
	file, err := os.Open(s.getFullPath(base, fileName))
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, ErrIsDir
	}
	data := make([]byte, info.Size())
	_, err = file.Read(data)
	return data, err
}

// save space stat
func (s *Space) save() error {
	return os.WriteFile(
		s.getStatPath(),
		[]byte(fmt.Sprintf("%d,%d", s.capacity, s.occupy)),
		0755,
	)
}

func (s *Space) Close(calcSize ...bool) error {
	if len(calcSize) == 1 && calcSize[0] {
		s.occupy, _ = s.GetSize("", "")
	}
	log.Println("close space", s.spaceKey)
	return s.save()
}

// unit can be "B", "KB", "MB", "GB" or just leave it blank
func (s *Space) Occupy(unit ...string) float64 {
	if s == nil {
		panic("basicFileSystem is nil")
	}
	if len(unit) == 0 {
		return float64(s.occupy)
	}
	switch unit[0] {
	case "B":
		return float64(s.occupy)
	case "KB":
		return float64(s.occupy) / 1024
	case "MB":
		return float64(s.occupy) / 1024 / 1024
	case "GB":
		return float64(s.occupy) / 1024 / 1024 / 1024
	default:
		return float64(s.occupy)
	}
}

// Get "this" size, "this" can be a file or a dir
func (s *Space) GetSize(base, target string) (Byte, error) {
	var size Byte
	err := filepath.WalkDir(s.getFullPath(base, target), func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return fmt.Errorf("path %s not exist", path)
		}
		if d.IsDir() {
			return nil
		}
		info, _ := d.Info()
		size += Byte(info.Size())
		return nil
	})
	return size, err
}

func (s *Space) Cap() Byte {
	return s.capacity
}

func (s *Space) ModifyCap(cap Byte) error {
	if cap < s.occupy {
		return ErrSpaceInternal
	}
	s.capacity = cap
	return s.save()
}

// func (s *Space) willFull(delta int64) bool {
// 	return s.occupy+delta > s.capacity
// }

func (s *Space) getFullPath(base, target string) string {
	if strings.Contains(base, BASE_DIR) {
		return filepath.Join(s.root, s.spaceKey, base, target)
	}
	return filepath.Join(s.root, s.spaceKey, s.base, base, target)
}

func (s *Space) getStatPath() string {
	return filepath.Join(s.root, s.spaceKey, STAT_FILE)
}
