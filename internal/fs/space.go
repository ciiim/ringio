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

// xxx/zzz/file.txt
func (s *Space) Store(fullpath string, data []byte) (err error) {
	sep := strings.Split(fullpath, "/")
	if strings.Contains(sep[len(sep)-1], DIR_PERFIX) {
		sep[len(sep)-1] = strings.TrimLeft(sep[len(sep)-1], DIR_PERFIX)
		fullpath = strings.Join(sep, "/")
		err = s.MkDir(fullpath)
	} else {
		err = s.storeFile(fullpath, data, os.O_CREATE|os.O_WRONLY)
	}
	return err
}

func (s *Space) Get(fullpath string) (File, error) {
	stat, err := os.Stat(s.getFullPath(fullpath))
	if err != nil {
		return nil, err
	}
	if stat == nil {
		return nil, ErrFileNotFound
	}
	if stat.IsDir() {
		subDir, _ := s.getDir(fullpath)
		return TreeFile{
			data: nil,
			info: TreeFileInfo{
				NewFileInfo(stat.Name(), "", fullpath, stat.Size(), true, stat.ModTime()),
				DirEntryToSubList(subDir),
			},
		}, nil
	} else {
		return s.getFile(fullpath)
	}
}

func (s *Space) Delete(fullpath string) error {

	//TODO: 防止删除fullpath的上级目录
	size, err := s.GetSize(fullpath)
	if err != nil {
		return err
	}
	if s.occupy < size {
		return ErrInternal
	}
	s.occupy -= int64(size)

	return os.RemoveAll(s.getFullPath(fullpath))
}

func (s *Space) MkDir(fullpath string) error {

	//TODO: 防止创建到dir的上级目录内

	return os.Mkdir(s.getFullPath(fullpath), 0755)
}

func (s *Space) getDir(fullpath string) ([]fs.DirEntry, error) {

	//TODO: 防止访问dir的上级目录
	return os.ReadDir(s.getFullPath(fullpath))
}

func (s *Space) storeFile(fullpath string, data []byte, flag int) error {

	if s.willFull(int64(len(data))) {
		return ErrFull
	}
	file, err := os.OpenFile(s.getFullPath(fullpath), flag, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	info, _ := file.Stat()
	oldSize := info.Size()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	newSize := info.Size()
	s.occupy += Byte(newSize - oldSize)

	return nil
}

func (s *Space) getFile(fullpath string) (File, error) {
	file, err := os.Open(s.getFullPath(fullpath))
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	// TODO: 补全hash和path
	bfi := NewFileInfo(info.Name(), "", fullpath, info.Size(), info.IsDir())
	data := make([]byte, info.Size())
	_, err = file.Read(data)
	return TreeFile{
		data: data,
		info: TreeFileInfo{
			bfi,
			nil,
		},
	}, err
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
		s.occupy, _ = s.GetSize("")
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
func (s *Space) GetSize(fullpath string) (Byte, error) {
	var size Byte
	err := filepath.WalkDir(s.getFullPath(fullpath), func(path string, d fs.DirEntry, err error) error {
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

func (s *Space) willFull(delta int64) bool {
	return s.occupy+delta > s.capacity
}

func (s *Space) getFullPath(fullpath string) string {
	if strings.Contains(fullpath, BASE_DIR) {
		return filepath.Join(s.root, s.spaceKey, fullpath)
	}
	return filepath.Join(s.root, s.spaceKey, s.base, fullpath)
}

func (s *Space) getStatPath() string {
	return filepath.Join(s.root, s.spaceKey, STAT_FILE)
}
