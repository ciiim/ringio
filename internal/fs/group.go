package fs

import (
	"errors"
	"io"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

const (
	GroupFSLimit int  = 1
	BLOCK_SIZE   Byte = 1024 * 1024 * 4 // 4MB
)

/*
Group is a group of file systems.

It contains a front system and a list of distributed file systems.

front system is used to store the meta data of the file.

distributed file systems are used to store the blocks of the file.
*/
type Group struct {
	groupName string

	/*
		It will store the meta data of the file
	*/
	FrontSystem DistributeFileSystem

	/*
		A list of File System

		Only can use one FileSystem just now.
		XXX: support redundancy in the future.
	*/
	StoreSystems []DistributeFileSystem
}

func NewGroup(groupName string, frontSystem DistributeFileSystem) *Group {
	return &Group{
		groupName:    groupName,
		StoreSystems: make([]DistributeFileSystem, 0, 10),
		FrontSystem:  frontSystem,
	}
}

func (g *Group) SetFrontSystem(fs DistributeFileSystem) {
	if g.FrontSystem != nil {
		log.Println("[Group] DO NOT set front system again.")
		return
	}
	g.FrontSystem = fs
}

func (g *Group) UseFS(fs ...DistributeFileSystem) {
	if len(g.StoreSystems)+len(fs) > GroupFSLimit {
		log.Println("[Group] Reached the limit")
		return
	}
	g.StoreSystems = append(g.StoreSystems, fs...)
}

func (g *Group) Serve() {
	for _, fs := range g.StoreSystems {
		go fs.Serve()
	}
	g.FrontSystem.Serve()
}

//User method field

func (g *Group) NewBorad(spaceKey string) error {
	return g.FrontSystem.Store(spaceKey, NEW_SPACE, nil)
}

func (g *Group) StoreFile(spaceKey, filehash, basePath, filename string, blocksStream io.ReadCloser, blocks []Fileblock) error {
	if blocksStream == nil {
		return errors.New("blocksStream is nil")
	}

	//calculate the file size
	var filesize int64
	for _, block := range blocks {
		filesize += block.Size
	}

	// generate the metadata
	metadata := newMetaData(filename, filehash, filesize, time.Now(), blocks)
	metadataBytes := marshalMetaData(metadata)

	//save metadata
	err := g.FrontSystem.Store(spaceKey, filepath.Join(basePath, filename+META_FILE_SUFFIX), metadataBytes)
	if err != nil {
		return err
	}

	//save blocks
	var wg sync.WaitGroup
	wg.Add(len(g.StoreSystems))
	defer blocksStream.Close()
	log.Println("[Group] Start to store blocks")
	//TODO
	wg.Wait()
	return nil
}

func (g *Group) Delete(spaceKey, fullpath string) error {

	// You can see the format definition in dtreefs.go -> Delete Function
	delString := filepath.Join(spaceKey, fullpath)
	meta, err := g.GetMetaData(delString)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(meta.Blocks))
	for _, block := range meta.Blocks {
		go g.DeleteBlock(block, &wg)
	}
	wg.Wait()
	return g.DeleteMetaData(delString)
}

func (g *Group) Mkdir(spaceKey, basePath, dirName string) error {

	//Add DIR_PERFIX for FileSystem to Specify the dir
	//But the real dir name that store in the front system is the dirName.
	realDirString := filepath.Join(basePath, DIR_PERFIX+dirName)
	log.Println("[Group] Mkdir:", realDirString)
	return g.FrontSystem.Store(spaceKey, realDirString, nil)
}

func (g *Group) GetDir(spaceKey, basePath, dirName string) (File, error) {

	// You can see the format definition in dtreefs.go -> Get Function
	getString := filepath.Join(spaceKey, basePath, dirName)
	log.Printf("[Group] Get dir:%s\n", getString)
	return g.FrontSystem.Get(getString)
}

/*
it will close the front system and all the file systems in the list.

Return the last error,

other error will be logged.
*/
func (g *Group) Close() error {
	err := g.FrontSystem.Close()
	if err != nil {
		log.Println("[Group] Close front system error:", err)
	}
	for _, fs := range g.StoreSystems {
		err = fs.Close()
		if err != nil {
			log.Println("[Group] Close file system error:", err)
		}
	}
	return err
}

/*
key - format: <spaceKey>/<filefullpath>
*/
func (g *Group) GetMetaData(key string) (Metadata, error) {
	//get metadata
	metadataFile, err := g.FrontSystem.Get(key + META_FILE_SUFFIX)
	if err != nil {
		return Metadata{}, err
	}
	metadataBytes := metadataFile.Data()

	//read metadata
	var meta Metadata
	readMetaDataByBytes(metadataBytes, &meta)

	return meta, nil
}

func (g *Group) DeleteMetaData(key string) error {
	return g.FrontSystem.Delete(key + META_FILE_SUFFIX)
}

func (g *Group) GetBlockData(blockInfo Fileblock) ([]byte, error) {
	var err error
	for _, fs := range g.StoreSystems {
		file, err := fs.Get(blockInfo.Hash)
		if err == nil && int64(len(file.Data())) == blockInfo.Size {
			return file.Data(), nil
		}
	}
	return nil, err
}

func (g *Group) DeleteBlock(blockInfo Fileblock, wg *sync.WaitGroup) error {
	var err error
	for _, fs := range g.StoreSystems {
		err = fs.Delete(blockInfo.Hash)
		if err == nil {
			return nil
		}
	}
	wg.Done()
	return err
}

func (g *Group) PeerList() []DPeerInfo {
	peers := make([]DPeerInfo, len(g.FrontSystem.Peer().PList()))

	return peers
}

/*
pi - one of the peer info in the group
*/
func (g *Group) Join(pi peers.PeerInfo) error {

	//boradcast to group and get all peers of the group

	err := g.FrontSystem.Peer().PSync(pi, peers.P_ACTION_JOIN)
	if err != nil {
		return err
	}
	for _, fs := range g.StoreSystems {
		err = fs.Peer().PSync(pi, peers.P_ACTION_JOIN)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Group) Quit() {
	g.FrontSystem.Peer().PSync(g.FrontSystem.Peer().Info(), peers.P_ACTION_QUIT)
	for _, fs := range g.StoreSystems {
		fs.Peer().PSync(g.FrontSystem.Peer().Info(), peers.P_ACTION_QUIT)
	}
}

func (g *Group) SyncPeer(pi peers.PeerInfo, action peers.PeerActionType) error {
	err := g.FrontSystem.Peer().PSync(pi, action)
	if err != nil {
		return err
	}
	for _, fs := range g.StoreSystems {
		err = fs.Peer().PSync(pi, action)
		if err != nil {
			return err
		}
	}
	return nil
}
