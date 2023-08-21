package fs

import (
	"errors"
	"io/fs"
	"log"
	"strings"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

// Distributed Tree File System.
// Implement FileSystem interface
type DTFS struct {
	*treeFS
	self DPeer
}

type DTreeFile struct {
	data []byte
	info DTreeFileInfo
}
type DTreeFileInfo struct {
	TreeFileInfo
	DPeerInfo
	subDir []fs.DirEntry
}

// make sure DTFS implement FileSystem interface
var _ DistributeFileSystem = (*DTFS)(nil)

func NewDTFS(self DPeer, rootPath string) *DTFS {
	dtfs := &DTFS{
		self:   self,
		treeFS: NewTreeFS(rootPath),
	}
	return dtfs

}

func (dt *DTFS) AddPeer(pi ...peers.PeerInfo) {
	dt.self.PAdd(pi...)
}

func (dt *DTFS) PickPeer(key string) peers.PeerInfo {
	return dt.self.Pick(key)
}

/*
Store a file or create a dir.

key - sapce key.

name - file or dir full path. dir should have DIR_PERFIX.

value - file content. dir should be nil.
*/
func (dt *DTFS) Store(key, name string, value []byte) error {
	pi := dt.PickPeer(key)
	if pi == nil {
		return peers.ErrPeerNotFound
	}
	if pi.Equal(dt.self.info) {
		return dt.storeLocally(key, name, value)
	}

	return dt.self.Put(pi, key, name, value).Err
}

// key - format: spacekey/fullpath
func (dt *DTFS) Get(key string) (File, error) {
	spacekey, path := splitKey(key)
	pi := dt.PickPeer(key)
	if pi == nil {
		return nil, peers.ErrPeerNotFound
	}
	if pi.Equal(dt.self.info) {
		df, err := dt.getLocally(spacekey, path)
		if errors.Is(err, ErrFileNotFound) {
			return dt.recoverFile(key)
		} else {
			return df, err
		}
	}
	resp := dt.self.Get(pi, key)
	if resp.Err != nil {
		return nil, resp.Err
	}

	df := DTreeFile{
		data: resp.Data,
		info: resp.Info.(DTreeFileInfo),
	}
	return df, resp.Err
}

func (dt *DTFS) Delete(key string) error {
	pi := dt.PickPeer(key)
	if pi == nil {
		return peers.ErrPeerNotFound
	}
	if pi.Equal(dt.self.info) {
		return dt.deleteLocally(key)
	}
	return dt.self.Delete(pi, key).Err
}

func (dt *DTFS) Close() (err error) {
	for _, s := range dt.openSpaces {
		if e := s.Close(); err != nil {
			err = e
			log.Println("[DTFS] Close space error:", err)
		}

	}
	return err
}

func (dt *DTFS) storeLocally(spacekey string, fullpath string, data []byte) error {
	space := dt.GetSpace(spacekey)
	if space == nil {
		return ErrSpaceNotFound
	}
	return space.Store(fullpath, data)
}

func (dt *DTFS) getLocally(key string, fullpath string) (File, error) {
	spacekey, path := splitKey(key)
	space := dt.GetSpace(spacekey)
	if space == nil {
		return nil, ErrSpaceNotFound
	}
	f, err := space.Get(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

/*
key - format: sapcekey/fullpath

e.g. "user/1/2/3" -> "user", "1/2/3",
will delete "1/2/3" from "user" space.
*/
func (dt *DTFS) deleteLocally(key string) error {
	spacekey, path := splitKey(key)
	space := dt.GetSpace(spacekey)
	if space == nil {
		return ErrSpaceNotFound
	}
	return space.Delete(path)
}

func (dt *DTFS) Set(opt any) error {
	//TODO: set options
	return nil
}

func (dt *DTFS) Peer() peers.Peer {
	return dt.self
}

func (dt *DTFS) recoverFile(key string) (File, error) {
	nextInfo := dt.Peer().PNext(key)
	if nextInfo == nil {
		return nil, peers.ErrPeerNotFound
	}
	if nextInfo.Equal(dt.self.Info()) {
		return nil, ErrFileNotFound
	}
	// Get file info from next peer
	resp := dt.self.Get(nextInfo, key)
	if resp.Err == nil {
		// delete the wrong local file
		dt.self.Delete(nextInfo, key)
		return DTreeFile{
			data: resp.Data,
			info: resp.Info.(DTreeFileInfo),
		}, resp.Err
	}
	return nil, resp.Err
}

func splitKey(key string) (spacekey, path string) {
	temp := strings.SplitN(key, "/", 2)
	return temp[0], temp[1]
}

func (df DTreeFile) Data() []byte {
	return df.data
}

func (df DTreeFile) Stat() FileInfo {
	return df.info
}

func (dfi DTreeFileInfo) PeerInfo() peers.PeerInfo {
	return dfi.DPeerInfo
}

func (dfi DTreeFileInfo) SubDir() []fs.DirEntry {
	return dfi.subDir
}

func (dt *DTFS) Serve() {
	log.Println("[DTFS] Serve on ", dt.self.PAddr())
	newRpcServer(dt).run(FRONT_PORT)
}
