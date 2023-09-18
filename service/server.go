package service

import (
	"github.com/ciiim/cloudborad/internal/fs/remote"
)

func (s *Service) JoinCluster(name, addr string) error {
	return s.fileServer.JoinCluster(name, addr)
}

func (s *Service) QuitCluster() {
	s.fileServer.QuitCluster()
}

func (s *Service) CloseServer() {
	s.fileServer.Group.Close()
}

func (s *Service) ServerInfo() (string, string) {
	return s.fileServer.ServerInfo()
}

func (s *Service) GetClusterList() []remote.DPeerInfo {
	return remote.PeerInfoListToDpeerInfoList(s.fileServer.Group.FrontSystem.Peer().PList())
}
