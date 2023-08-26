package fs

import (
	"log"
	"sync"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

const (
	GroupFSLimit int  = 1
	BLOCK_SIZE   Byte = 1024 * 1024 * 5 // 5MB
)

//TODO: 事务(transaction)系统，提供事务接口，支持事务回滚，实现文件下载和上传的事务
//TODO: 秒传模块，相同Hash的文件秒传
//TODO: 节点强一致性，保证节点信息一致性
/*
Group 是一个文件系统的集合，包含一个前端文件系统和一个后端文件系统。

前端文件系统负责文件的元数据管理，后端文件系统负责文件的存储。
*/
type Group struct {
	groupName string

	/*
		It will store the meta data of the file
	*/
	FrontSystem TreeDFileSystemI

	/*
		A list of File System

		Only can use one FileSystem just now.
		XXX: support redundancy in the future.
	*/
	StoreSystem HashDFileSystemI
}

func NewGroup(groupName string, frontSystem TreeDFileSystemI, storeSystem HashDFileSystemI) *Group {
	return &Group{
		groupName:   groupName,
		StoreSystem: storeSystem,
		FrontSystem: frontSystem,
	}
}
func (g *Group) Serve() {
	go g.StoreSystem.Serve()
	g.FrontSystem.Serve()
}

//User method field

func (g *Group) NewBorad(spaceKey string) error {
	return g.FrontSystem.NewSpace(spaceKey, GB)
}

func (g *Group) DeleteFile(spaceKey, base, name string) error {

	// You can see the format definition in dtreefs.go -> Delete Function
	meta, err := g.GetMetaData(spaceKey, base, name)
	if err != nil {
		return err
	}
	metadata := &Metadata{}
	UnmarshalMetaData(meta, metadata)
	var wg sync.WaitGroup
	wg.Add(len(metadata.Blocks))
	for _, block := range metadata.Blocks {
		go g.DeleteBlock(block, &wg)
	}
	wg.Wait()
	return g.DeleteMetaData(spaceKey, base, name)
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
	err = g.StoreSystem.Close()
	return err
}

/*
key - format: <spaceKey>/<filefullpath>
*/
func (g *Group) GetMetaData(space, base, name string) ([]byte, error) {
	//get metadata
	metadata, err := g.FrontSystem.GetMetadata(space, base, name)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (g *Group) DeleteMetaData(space, base, name string) error {
	return g.FrontSystem.DeleteMetadata(space, base, name+META_FILE_SUFFIX)
}

func (g *Group) GetBlockData(blockInfo Fileblock) ([]byte, error) {
	var err error
	file, err := g.StoreSystem.Get(blockInfo.Hash)
	if err != nil {
		return file.Data(), nil
	}
	return nil, err
}

func (g *Group) DeleteBlock(blockInfo Fileblock, wg *sync.WaitGroup) error {
	err := g.StoreSystem.Delete(blockInfo.Hash)
	return err
}

//Peer method field

func (g *Group) AddPeer(name, addr string) {
	if name == "" || addr == "" {
		return
	}
	g.FrontSystem.Peer().PAdd(NewDPeerInfo(name, WithPort(addr, RPC_TDFS_PORT)))
	g.StoreSystem.Peer().PAdd(NewDPeerInfo(name, WithPort(addr, RPC_HDFS_PORT)))
}

func (g *Group) PeerList() []DPeerInfo {
	peers := make([]DPeerInfo, len(g.FrontSystem.Peer().PList()))
	return peers
}

/*
pi - one of the peer info in the group
*/
func (g *Group) Join(name, addr string) error {

	//boradcast to group and get all peers of the group

	frontDest := NewDPeerInfo(name, WithPort(addr, RPC_TDFS_PORT))
	storeDest := NewDPeerInfo(name, WithPort(addr, RPC_HDFS_PORT))

	//Join Cluster
	err := g.FrontSystem.Peer().PActionTo(peers.P_ACTION_JOIN, frontDest)
	if err != nil {
		return err
	}

	// Get List from cluster
	peerList, err := g.FrontSystem.Peer().GetPeerListFromPeer(frontDest)
	if err != nil {
		return err
	}

	//Add to peer map
	for _, peer := range peerList {
		_ = g.FrontSystem.Peer().PSync(peer, peers.P_ACTION_NEW)
	}

	//Join Cluster
	err = g.StoreSystem.Peer().PActionTo(peers.P_ACTION_JOIN, storeDest)
	if err != nil {
		return err
	}

	// Get List from cluster
	peerList, err = g.StoreSystem.Peer().GetPeerListFromPeer(storeDest)
	if err != nil {
		return err
	}

	//Add to peer map
	for _, peer := range peerList {
		_ = g.StoreSystem.Peer().PSync(peer, peers.P_ACTION_NEW)
	}

	return nil
}

func (g *Group) Quit() {
	g.FrontSystem.Peer().PSync(g.FrontSystem.Peer().Info(), peers.P_ACTION_QUIT)
	g.StoreSystem.Peer().PSync(g.FrontSystem.Peer().Info(), peers.P_ACTION_QUIT)
}

func (g *Group) SyncPeer(pi peers.PeerInfo, action peers.PeerActionType) error {
	err := g.FrontSystem.Peer().PSync(pi, action)
	if err != nil {
		return err
	}
	err = g.StoreSystem.Peer().PSync(pi, action)
	return err
}
