package server

import (
	"log"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/internal/fs/peers"
	"github.com/ciiim/cloudborad/internal/fs/remote"
)

const (
	OPTION_NO_FRONT ServerOptions = iota
	OPTION_NO_STORE
	OPTION_LOCAL
	OPTION_FUCK
)

type ServerOptions int

func (s *Server) handleOptions(options ...ServerOptions) {
	var (
		front remote.TreeDFileSystemI = remote.DefaultTreeDFileSystem{}
		store remote.HashDFileSystemI = remote.DefaultHashDFileSystem{}
		ps    peers.Peer              = peers.DefaultPeer{}
	)

	var (
		bfont  bool = true
		bstore bool = true
		bps    bool = true
	)
	for _, option := range options {
		switch option {
		case OPTION_FUCK:
			log.Fatalf("[Server] Option: You Fucked Server.")
		case OPTION_NO_FRONT:
			log.Println("[Server] Disabled Front FileSystem.")
			bfont = false
		case OPTION_NO_STORE:
			log.Println("[Server] Disabled Store FileSystem.")
			bstore = false
		case OPTION_LOCAL:
			log.Println("[Server] Disabled Peer Service.")
			bps = false
		}
	}
	if bps {
		ps = remote.NewDPeer("_fs_"+s.serverName, remote.WithPort(s._IP, s._Port), 20, nil)
	}
	if bfont {
		t := remote.NewTreeDFileSystem("./_fs_/front_" + s.serverName)
		t.SetPeerService(ps)
		front = t
	}
	if bstore {
		t := remote.NewDFS("./_fs_/store_"+s.serverName, 50*fs.GB, nil)
		t.SetPeerService(ps)
		store = t
	}

	s.Group = remote.NewGroup(s.serverName, ps, front, store)
}
