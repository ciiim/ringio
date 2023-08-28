package server

import (
	"log"

	"github.com/ciiim/cloudborad/internal/fs"
)

func (s *Server) ServerInfo() (string, string) {
	return s.serverName, s._IP
}

func (s *Server) AddPeer(name, addr string) {
	s.Group.AddPeer(name, addr)
}

func (s *Server) Join(peerName, peerAddr string) error {
	err := s.JoinCluster(peerName, peerAddr)
	if err != nil {
		return err
	}
	log.Println("[Server] Join cluster success")
	return nil
}

func (s *Server) Quit() {
	s.QuitCluster()
}

func (s *Server) Close() error {
	s.Quit()
	return s.Group.Close()
}

func (s *Server) DebugOn() {
	fs.DebugOn()
	log.Println("[WARNING] DEBUG MODE ON")
}
