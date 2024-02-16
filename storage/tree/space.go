package tree

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ciiim/cloudborad/storage/types"
)

var (
	ErrSpaceExist    = fmt.Errorf("space exist")
	ErrSpaceFull     = fmt.Errorf("space full")
	ErrSpaceNotFound = fmt.Errorf("space not found")
	ErrSpaceInternal = fmt.Errorf("space internal error")
	ErrIsDir         = fmt.Errorf("is dir error")
)

type Space struct {
	root string

	spaceMu sync.Mutex

	// /treeFS.rootPath/spaceKey
	spaceKey string
	base     string
	capacity types.Byte
	occupy   types.Byte
}

func (s *Space) storeMetaData(base, fileName string, metadata []byte) (err error) {

	return s.i_storeMetadata(base, fileName, metadata, os.O_CREATE|os.O_WRONLY|os.O_EXCL)

}

func (s *Space) getMetadata(base, fileName string) ([]byte, error) {
	return s.i_getMetadata(base, fileName)
}

func (s *Space) deleteMetaData(base, fileName string) error {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()
	size, err := s.GetSize(base, fileName)
	if err != nil {
		return err
	}
	if s.occupy < size {
		return errors.New("occupy < size")
	}

	s.occupy -= size

	return os.Remove(s.getFullPath(base, fileName))
}

func (s *Space) makeDir(base, fileName string) error {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()
	return os.Mkdir(s.getFullPath(base, fileName), 0755)
}

func (s *Space) renameDir(base, fileName, newName string) error {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()
	return os.Rename(s.getFullPath(base, fileName), s.getFullPath(base, newName))
}

func (s *Space) deleteDir(base, fileName string) error {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()
	return os.RemoveAll(s.getFullPath(base, fileName))
}

func (s *Space) getDir(base, fileName string) (string, []fs.DirEntry, error) {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()
	fullpath := s.getFullPath(base, fileName)
	dirs, err := os.ReadDir(fullpath)

	return fullpath, dirs, err
}
func (s *Space) i_storeMetadata(base, fileName string, metadata []byte, flag int) error {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()

	// 加元数据后缀
	fileName = fileName + META_FILE_SUFFIX

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

	s.occupy += (newSize - oldSize)

	return nil
}

func (s *Space) i_getMetadata(base, fileName string) ([]byte, error) {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()

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
	data := make([]byte, info.Size()+1)
	_, err = file.Read(data)
	return data, err
}

func (s *Space) metadataExist(base, fileName string) bool {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()

	_, err := os.Stat(s.getFullPath(base, fileName))
	return !os.IsNotExist(err)
}

// save space stat
func (s *Space) save() error {
	return os.WriteFile(
		s.getStatPath(),
		[]byte(fmt.Sprintf("%d,%d", s.capacity, s.occupy)),
		0755,
	)
}

func (s *Space) Close() error {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()
	return s.save()
}

func (s *Space) Occupy() types.Byte {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()

	return s.occupy
}

// Get "this" size, "this" can be a file or a dir
func (s *Space) GetSize(base, target string) (types.Byte, error) {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()

	var size types.Byte
	err := filepath.WalkDir(s.getFullPath(base, target), func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return fmt.Errorf("path %s not exist", path)
		}
		if d.IsDir() {
			return nil
		}
		info, _ := d.Info()

		size += info.Size()
		return nil
	})
	return size, err
}

func (s *Space) Cap() int64 {
	s.spaceMu.Lock()
	defer s.spaceMu.Unlock()

	return s.capacity
}

func (s *Space) getFullPath(base, target string) string {
	if strings.Contains(base, BASE_DIR) {
		return filepath.Join(s.root, s.spaceKey, base, target)
	}
	return filepath.Join(s.root, s.spaceKey, s.base, base, target)
}

func (s *Space) getStatPath() string {
	return filepath.Join(s.root, s.spaceKey, STAT_FILE)
}
