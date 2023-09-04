package server

import (
	"log"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/internal/fs/peers"
)

func (s *Server) ServerInfo() (string, string) {
	return s.serverName, s._IP
}

func (s *Server) AddPeer(name, addr string) {
	s.Group.AddPeer(name, addr)
}

func (s *Server) JoinCluster(name, addr string) error {
	//boradcast to group and get all peers of the group

	frontDest := fs.NewDPeerInfo(name, fs.WithPort(addr, fs.RPC_TDFS_PORT))
	storeDest := fs.NewDPeerInfo(name, fs.WithPort(addr, fs.RPC_HDFS_PORT))

	//Join Cluster
	err := s.Group.FrontSystem.Peer().PActionTo(peers.P_ACTION_JOIN, frontDest)
	if err != nil {
		return err
	}

	// Get List from cluster
	peerList, err := s.Group.FrontSystem.Peer().GetPeerListFromPeer(frontDest)
	if err != nil {
		return err
	}

	//Add to peer map
	for _, peer := range peerList {
		_ = s.Group.FrontSystem.Peer().PSync(peer, peers.P_ACTION_NEW)
	}

	//Join Cluster
	err = s.Group.StoreSystem.Peer().PActionTo(peers.P_ACTION_JOIN, storeDest)
	if err != nil {
		return err
	}

	// Get List from cluster
	peerList, err = s.Group.StoreSystem.Peer().GetPeerListFromPeer(storeDest)
	if err != nil {
		return err
	}

	//Add to peer map
	for _, peer := range peerList {
		_ = s.Group.StoreSystem.Peer().PSync(peer, peers.P_ACTION_NEW)
	}

	return nil
}

func (s *Server) QuitCluster() error {
	list := s.Group.FrontSystem.Peer().PList()

	err := s.Group.FrontSystem.Peer().PActionTo(peers.P_ACTION_QUIT, list...)
	if err != nil {
		return err
	}
	return s.Group.StoreSystem.Peer().PActionTo(peers.P_ACTION_QUIT, list...)

}

func (s *Server) DebugOn() {
	fs.DebugOn()
	log.Println("[WARNING] DEBUG MODE ON")
}
