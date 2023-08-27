package fs

import (
	"log"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

const (
	GroupFSLimit int  = 1
	BLOCK_SIZE   Byte = 1024 * 1024 * 5 // 5MB
)

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
