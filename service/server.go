package service

import "github.com/ciiim/cloudborad/internal/fs"

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

func (s *Service) GetClusterList() []fs.DPeerInfo {
	return fs.PeerInfoListToDpeerInfoList(s.fileServer.Group.FrontSystem.Peer().PList())
}
