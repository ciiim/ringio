package ringio

import (
	"log"
	"net"

	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/types"
	"google.golang.org/grpc"
)

const (
	DefaultBufferSize types.Byte = 4 * types.MB
)

type rpcServer struct {
	s *grpc.Server

	defaultBufferSize int64

	hcs         IDHashChunkSystem
	tfs         ITreeDFileSystem
	nodeService *node.NodeServiceRO
	fspb.UnimplementedHashChunkSystemServiceServer
	fspb.UnimplementedTreeFileSystemServiceServer
}

func newRPCFSServer(ns *node.NodeServiceRO, hfs IDHashChunkSystem, tfs ITreeDFileSystem) *rpcServer {
	return &rpcServer{
		nodeService:       ns,
		hcs:               hfs,
		tfs:               tfs,
		s:                 grpc.NewServer(),
		defaultBufferSize: DefaultBufferSize,
	}
}

func (r *rpcServer) serve(port string) {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("[RPC Server] Listen error:", err)
		return
	}
	//TODO: 注册节点服务
	fspb.RegisterHashChunkSystemServiceServer(r.s, r)
	fspb.RegisterTreeFileSystemServiceServer(r.s, r)
	err = r.s.Serve(l)
	if err != nil {
		log.Println("[RPC Server] Server shutdown:", err)
		return
	}
}
