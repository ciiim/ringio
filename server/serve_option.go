package server

import (
	"log"

	"github.com/ciiim/cloudborad/internal/dfs"
	"github.com/ciiim/cloudborad/internal/dfs/peers"
	"github.com/ciiim/cloudborad/internal/fs"
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
		front dfs.TreeDFileSystemI = dfs.DefaultTreeDFileSystem{}
		store dfs.HashDFileSystemI = dfs.DefaultHashDFileSystem{}
		ps    peers.Peer           = peers.DefaultPeer{}
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
		ps = dfs.NewDPeer("_fs_"+s.serverName, dfs.WithPort(s._IP, s._Port), 20, nil, dfs.DefaultSyncSettings) //FIXME:不要使用默认配置
	}
	if bfont {
		t := dfs.NewTreeDFileSystem("./_fs_/front_" + s.serverName)
		t.SetPeerService(ps)
		front = t
	}
	if bstore {
		t := dfs.NewDFS("./_fs_/store_"+s.serverName, 50*fs.GB, nil)
		t.SetPeer(ps)
		store = t
	}

	s.Group = dfs.NewGroup(s.serverName, ps, front, store)
}
