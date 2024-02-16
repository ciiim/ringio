package tree

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/ciiim/cloudborad/storage/types"
)

const (
	SPACE_DEFAULT_CAP = 1024 * 1024 * 100 // 100MB
)

type DirEntry = fs.DirEntry

type TreeFileSystem struct {
	rootPath string
}

var _ ITreeFileSystem = (*TreeFileSystem)(nil)

const (
	DIR_PERFIX = "__DIR__"
	STAT_FILE  = "__STAT__"
	BASE_DIR   = "__BASE__"
)

func NewTreeFileSystem(rootPath string) *TreeFileSystem {
	err := os.MkdirAll(rootPath, os.ModePerm)
	if err != nil {
		return nil
	}
	t := &TreeFileSystem{
		rootPath: rootPath,
	}
	return t
}

func (t *TreeFileSystem) NewLocalSpace(space string, cap types.Byte) error {

	if _, err := os.Stat(filepath.Join(t.rootPath, space, BASE_DIR)); err == nil {
		return ErrSpaceExist
	}

	err := os.Mkdir(filepath.Join(t.rootPath, space), 0755)
	if err != nil {
		return err
	}
	err = os.Mkdir(filepath.Join(t.rootPath, space, BASE_DIR), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(t.rootPath, space, STAT_FILE))
	if err != nil {
		return err
	}

	_, err = file.WriteString(fmt.Sprintf("%d,0", cap))

	return err
}

func (t *TreeFileSystem) GetLocalSpace(space string) *Space {
	file, err := os.Open(filepath.Join(t.rootPath, space, STAT_FILE))
	if err != nil {
		log.Println("[Space] Lack of stat file", err)
		return nil
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return nil
	}
	capANDoccupy := strings.Split(string(b), ",")
	if len(capANDoccupy) != 2 {
		log.Println("[Space] stat file error")
		return nil
	}
	cap, _ := strconv.ParseInt(capANDoccupy[0], 10, 64)
	occupy, _ := strconv.ParseInt(capANDoccupy[1], 10, 64)
	s := &Space{
		root:     t.rootPath,
		spaceKey: space,
		base:     BASE_DIR,
		spaceMu:  sync.Mutex{},
		capacity: cap,
		occupy:   occupy,
	}
	return s
}

func (t *TreeFileSystem) DeleteLocalSpace(space string) error {
	if _, err := os.Stat(filepath.Join(t.rootPath, space, BASE_DIR)); err != nil {
		return ErrSpaceNotFound
	}
	return os.RemoveAll(filepath.Join(t.rootPath, space))
}

func (t *TreeFileSystem) MakeDir(space, base, name string) error {
	sp := t.GetLocalSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.makeDir(base, name)

}

func (t *TreeFileSystem) RenameDir(space, base, name, newName string) error {
	sp := t.GetLocalSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.renameDir(base, name, newName)

}

func (t *TreeFileSystem) DeleteDir(space, base, name string) error {
	sp := t.GetLocalSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.deleteDir(base, name)

}

func (t *TreeFileSystem) GetDirSub(space, base, name string) ([]*SubInfo, error) {
	sp := t.GetLocalSpace(space)
	if sp == nil {
		return nil, ErrSpaceNotFound
	}
	dirPath, des, err := sp.getDir(base, name)
	if err != nil {
		return nil, err
	}
	return DirEntryToSubInfo(dirPath, des), nil

}

func (t *TreeFileSystem) GetMetadata(space, base, name string) ([]byte, error) {
	sp := t.GetLocalSpace(space)
	if sp == nil {
		return nil, ErrSpaceNotFound
	}
	return sp.getMetadata(base, name)

}

func (t *TreeFileSystem) PutMetadata(space, base, name string, fileHash []byte, metadata []byte) error {
	sp := t.GetLocalSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}

	return sp.storeMetaData(base, name, metadata)

}
func (t *TreeFileSystem) DeleteMetadata(space, base, name string, hash []byte) error {
	sp := t.GetLocalSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.deleteMetaData(base, name)
}

func (t *TreeFileSystem) Close() error {
	return nil
}
