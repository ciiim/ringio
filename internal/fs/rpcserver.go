package fs

import (
	"log"
	"net"

	"github.com/ciiim/cloudborad/internal/fs/fspb"
	"github.com/ciiim/cloudborad/internal/fs/peers"
	"google.golang.org/grpc"
)

type rpcFSServer struct {
	s *grpc.Server

	hfs         HashDFileSystemI
	tfs         TreeDFileSystemI
	peerService peers.Peer
	fspb.UnimplementedPeerServiceServer
	fspb.UnimplementedHashFileSystemServiceServer
	fspb.UnimplementedTreeFileSystemServiceServer
}

func newRPCFSServer(ps peers.Peer, hfs HashDFileSystemI, tfs TreeDFileSystemI) *rpcFSServer {
	return &rpcFSServer{
		peerService: ps,
		hfs:         hfs,
		tfs:         tfs,
		s:           grpc.NewServer(),
	}
}

func (r *rpcFSServer) serve(port string) {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("[RPC Server] Listen error:", err)
		return
	}
	fspb.RegisterPeerServiceServer(r.s, r)
	fspb.RegisterHashFileSystemServiceServer(r.s, r)
	fspb.RegisterTreeFileSystemServiceServer(r.s, r)
	err = r.s.Serve(l)
	if err != nil {
		log.Println("[RPC Server] Server shutdown:", err)
		return
	}
}
