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

	rpcServer *rpcFSServer

	PeerService peers.Peer

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

func NewGroup(groupName string, peerService peers.Peer, frontSystem TreeDFileSystemI, storeSystem HashDFileSystemI) *Group {
	if frontSystem == nil {
		log.Println("[Group Warn] Empty front system")
	}
	if storeSystem == nil {
		log.Println("[Group Warn] Empty store system")
	}

	group := &Group{
		groupName:   groupName,
		StoreSystem: storeSystem,
		FrontSystem: frontSystem,
		PeerService: peerService,
		rpcServer:   newRPCFSServer(peerService, storeSystem, frontSystem),
	}
	return group
}

func (g *Group) Serve() {
	if g.PeerService == nil {
		log.Println("[Group] No peer service found")
		return
	}
	log.Printf("[Group] Peer service serve <%s> on port <%s>", g.groupName, g.PeerService.PAddr().Port())
	g.rpcServer.serve(g.PeerService.PAddr().Port())
}

/*
it will close the front system and all the file systems in the list.

if this server join a cluster, it will send a OFFLINE message to the cluster.

Return the last error,

other error will be logged.
*/
func (g *Group) Close() error {
	err := g.PeerService.PActionTo(peers.P_ACTION_OFFLINE, g.PeerService.PList()...)
	if err != nil {
		log.Println("[Group] Send offline message error:", err)
	}

	err = g.FrontSystem.Close()
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
	g.FrontSystem.Peer().PAdd(NewDPeerInfo(name, addr))
	g.StoreSystem.Peer().PAdd(NewDPeerInfo(name, addr))
}

func (g *Group) PeerList() []DPeerInfo {
	peers := make([]DPeerInfo, len(g.FrontSystem.Peer().PList()))
	return peers
}

func (g *Group) SyncPeer(pi peers.PeerInfo, action peers.PeerActionType) error {
	err := g.FrontSystem.Peer().PSync(pi, action)
	if err != nil {
		return err
	}
	err = g.StoreSystem.Peer().PSync(pi, action)
	return err
}
