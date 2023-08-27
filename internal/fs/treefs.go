package fs

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

const (
	SPACE_DEFAULT_CAP = 1024 * 1024 * 100 // 100MB
)

type DirEntry = fs.DirEntry

type treeFileSystem struct {
	rootPath string
}

type TreeFile struct {
	metadata Metadata
	info     TreeFileInfo
}

type TreeFileInfo struct {
	fileName string
	path     string
	size     int64
	isDir    bool
	modTime  time.Time
	subDir   []SubInfo
}

var _ TreeFileSystemI = (*treeFileSystem)(nil)
var _ TreeFileI = (*TreeFile)(nil)

const (
	DIR_PERFIX = "__DIR__"
	STAT_FILE  = ".__stat__"
	BASE_DIR   = "__BASE__"
)

func newTreeFileSystem(rootPath string) *treeFileSystem {
	err := os.MkdirAll(rootPath, os.ModePerm)
	if err != nil {
		panic("mkdir error:" + err.Error())
	}
	t := &treeFileSystem{
		rootPath: rootPath,
	}
	return t
}

func (t *treeFileSystem) NewSpace(space string, cap Byte) error {

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

func (t *treeFileSystem) GetSpace(space string) *Space {
	file, err := os.Open(filepath.Join(t.rootPath, space, STAT_FILE))
	if err != nil {
		log.Println("[Space] Lack of stat file", err)
		return nil
	}
	defer file.Close()
	stat, _ := file.Stat()
	temp := make([]byte, stat.Size())
	file.Read(temp)
	capANDoccupy := strings.Split(string(temp), ",")
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
		capacity: cap,
		occupy:   occupy,
	}
	return s
}

func (t *treeFileSystem) DeleteSpace(space string) error {
	if _, err := os.Stat(filepath.Join(t.rootPath, space, BASE_DIR)); err != nil {
		return ErrSpaceNotFound
	}
	return os.RemoveAll(filepath.Join(t.rootPath, space))
}

func (t *treeFileSystem) MakeDir(space, base, name string) error {
	sp := t.GetSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.makeDir(base, name)

}

func (t *treeFileSystem) RenameDir(space, base, name, newName string) error {
	sp := t.GetSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.renameDir(base, name, newName)

}

func (t *treeFileSystem) DeleteDir(space, base, name string) error {
	sp := t.GetSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.deleteDir(base, name)

}

func (t *treeFileSystem) GetDirSub(space, base, name string) ([]SubInfo, error) {
	sp := t.GetSpace(space)
	if sp == nil {
		return nil, ErrSpaceNotFound
	}
	des, err := sp.getDir(base, name)
	if err != nil {
		return nil, err
	}
	return DirEntryToSubInfo(des), nil

}

func (t *treeFileSystem) GetMetadata(space, base, name string) ([]byte, error) {
	sp := t.GetSpace(space)
	if sp == nil {
		return nil, ErrSpaceNotFound
	}
	return sp.getMetadata(base, name)

}

func (t *treeFileSystem) PutMetadata(space, base, name, fileHash string, data []byte) error {
	sp := t.GetSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}

	return sp.storeMetaData(base, name, data)

}
func (t *treeFileSystem) DeleteMetadata(space, base, name, hash string) error {
	sp := t.GetSpace(space)
	if sp == nil {
		return ErrSpaceNotFound
	}
	return sp.deleteMetaData(base, name)
}

func (t *treeFileSystem) Close() error {
	return nil
}

func (tf TreeFile) Metadata() []byte {
	data, _ := MarshalMetaData(&tf.metadata)
	return data
}

func (tf TreeFile) Stat() TreeFileInfoI {
	return tf.info
}

func (tfi TreeFileInfo) Name() string {
	return tfi.fileName
}

func (tfi TreeFileInfo) Size() int64 {
	return tfi.size
}

func (tfi TreeFileInfo) Path() string {
	return tfi.path
}

func (tfi TreeFileInfo) IsDir() bool {
	return tfi.isDir
}

func (tfi TreeFileInfo) ModTime() time.Time {
	return tfi.modTime
}

func (tfi TreeFileInfo) Sub() []SubInfo {
	if tfi.IsDir() {
		return tfi.subDir
	}
	return nil
}

func (tfi TreeFileInfo) PeerInfo() peers.PeerInfo {
	return nil
}
