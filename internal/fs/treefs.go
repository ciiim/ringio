package fs

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SPACE_DEFAULT_CAP = 1024 * 1024 * 100 // 100MB
)

type DirEntry = fs.DirEntry

type treeFS struct {
	rootPath string
	// capacity Byte
	// occupy   Byte
	openSpaces map[string]*Space
}

type TreeFile struct {
	data []byte
	info TreeFileInfo
}

type TreeFileInfo struct {
	BasicFileInfo
	subDir []SubInfo
}

var _ FileInfo = (*TreeFileInfo)(nil)

const (
	DIR_PERFIX = "__DIR__"
	STAT_FILE  = ".__stat__"
	BASE_DIR   = "__BASE__"
)

func NewTreeFS(rootPath string) *treeFS {
	t := &treeFS{
		rootPath:   rootPath,
		openSpaces: make(map[string]*Space, 16),
	}
	err := os.MkdirAll(rootPath, 0755)
	if err != nil {
		return nil
	}
	return t
}

func (t *treeFS) NewSpace(spaceKey string, cap Byte) (*Space, error) {

	if _, err := os.Stat(filepath.Join(t.rootPath, spaceKey, BASE_DIR)); err == nil {
		return t.GetSpace(spaceKey), ErrSpaceExist
	}

	err := os.Mkdir(filepath.Join(t.rootPath, spaceKey), 0755)
	if err != nil {
		return nil, err
	}
	err = os.Mkdir(filepath.Join(t.rootPath, spaceKey, BASE_DIR), 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(filepath.Join(t.rootPath, spaceKey, STAT_FILE))
	if err != nil {
		return nil, err
	}

	_, err = file.WriteString(fmt.Sprintf("%d,0", cap))

	return t.GetSpace(spaceKey), err
}

func (t *treeFS) GetSpace(spaceKey string) *Space {
	if s, ok := t.openSpaces[spaceKey]; ok {
		return s
	}
	file, err := os.Open(filepath.Join(t.rootPath, spaceKey, STAT_FILE))
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
		spaceKey: spaceKey,
		base:     BASE_DIR,
		capacity: cap,
		occupy:   occupy,
	}
	t.openSpaces[spaceKey] = s
	return s
}

func (t *treeFS) ModifySpace(spaceKey string, cap Byte) error {
	space, ok := t.openSpaces[spaceKey]
	if !ok {
		return ErrSpaceNotFound
	}
	return space.ModifyCap(cap)
}

func (t *treeFS) RemoveSpace(spaceKey string) error {

	//TODO: check if space is open

	return os.RemoveAll(filepath.Join(t.rootPath, spaceKey))
}

func (tf TreeFile) Data() []byte {
	return tf.data
}

func (tf TreeFile) Stat() FileInfo {
	return tf.info
}

func (tfi TreeFileInfo) SubDir() []SubInfo {
	if tfi.IsDir() {
		return tfi.subDir
	}
	return nil
}
