package server

import (
	"log"

	"github.com/ciiim/cloudborad/internal/fs/peers"

	"github.com/ciiim/cloudborad/internal/fs"
)

type Server struct {
	Group *fs.Group
}

func StartServer() {
	localServer := NewServer("default", "defaultServer0", "127.0.0.1")
	localServer.StartServer()
}

/*
ffs is the front file system

it must be a tree structure
*/
func NewServer(groupName, serverName, addr string) *Server {
	ffs := fs.NewDTFS(*fs.NewDPeer("front0_"+serverName+"_"+groupName, addr+":"+fs.FRONT_PORT, 20, nil), "./front0_"+serverName+"_"+groupName)
	sfs := fs.NewDFS(*fs.NewDPeer("store0_"+serverName+"_"+groupName, addr+":"+fs.FILE_STORE_PORT, 20, nil), "./store0_"+serverName+"_"+groupName, 1024*1024*1024, nil)
	if ffs == nil || sfs == nil {
		log.Fatal("New server failed")
	}
	server := &Server{
		Group: fs.NewGroup(groupName, ffs),
	}
	server.Group.UseFS(sfs)
	return server
}

func (s *Server) StartServer() {
	r := initRoute(s)
	go s.Group.Serve()
	r.Run(":8080")
}

func (s *Server) Join(peerName, peerAddr string) error {
	dest := fs.NewDPeerInfo(peerName, peerAddr)
	err := s.Group.FrontSystem.Peer().PActionTo(peers.P_ACTION_JOIN, dest)
	if err != nil {
		return err
	}
	for _, fs := range s.Group.StoreSystems {
		err = fs.Peer().PActionTo(peers.P_ACTION_JOIN, dest)
		if err != nil {
			return err
		}
	}
	log.Println("[Server] Join group success")
	return nil
}

func (s *Server) Quit() {
	s.Group.FrontSystem.Peer().PSync(s.Group.FrontSystem.Peer().Info(), peers.P_ACTION_QUIT)
	for _, fs := range s.Group.StoreSystems {
		fs.Peer().PSync(s.Group.FrontSystem.Peer().Info(), peers.P_ACTION_QUIT)
	}
}

func (s *Server) Close() error {
	s.Quit()
	return s.Group.Close()
}
