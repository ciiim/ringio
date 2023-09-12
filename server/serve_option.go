package server

import (
	"log"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/internal/fs/peers"
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
		front fs.TreeDFileSystemI = fs.DefaultTreeDFileSystem{}
		store fs.HashDFileSystemI = fs.DefaultHashDFileSystem{}
		ps    peers.Peer          = peers.DefaultPeer{}
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
	if bfont {
		front = fs.NewTreeDFileSystem("./_fs_/front_" + s.serverName)
	}
	if bstore {
		store = fs.NewDFS("./_fs_/store_"+s.serverName, 50*fs.GB, nil)
	}
	if bps {
		ps = fs.NewDPeer("_fs_"+s.serverName, fs.WithPort(s._IP, s._Port), 20, nil)
	}
	s.Group = fs.NewGroup(s.serverName, ps, front, store)
}
