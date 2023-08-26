package fs

import (
	"context"
	"log"
	"net"

	"github.com/ciiim/cloudborad/internal/fs/peers"

	"github.com/ciiim/cloudborad/internal/fs/fspb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type rpcHDFSServer struct {
	s *grpc.Server

	fs HashDFileSystemI
	fspb.UnimplementedPeerServiceServer
	fspb.UnimplementedHashFileSystemServiceServer
}

func newRPCHDFSServer(fs HashDFileSystemI) *rpcHDFSServer {
	return &rpcHDFSServer{
		fs: fs,
		s:  grpc.NewServer(),
	}
}

func (r *rpcHDFSServer) Get(ctx context.Context, key *fspb.Key) (*fspb.GetResponse, error) {
	file, err := r.fs.Get(key.Key)
	if err != nil {
		return nil, err
	}
	fi := file.Stat()

	return &fspb.GetResponse{
		Data: file.Data(),
		FileInfo: &fspb.HashFileInfo{
			FileName: fi.Name(),
			BasePath: fi.Path(),
			Hash:     fi.Hash(),
			Size:     fi.Size(),
		},
		PeerInfo: &fspb.PeerInfo{
			Name:   fi.PeerInfo().PName(),
			Addr:   fi.PeerInfo().PAddr(),
			Stat:   int64(fi.PeerInfo().PStat()),
			Action: int64(peers.P_ACTION_NONE),
		},
	}, nil
}

func (r *rpcHDFSServer) Put(ctx context.Context, req *fspb.PutRequest) (*fspb.Error, error) {
	if err := r.fs.Store(req.Key.Key, req.Filename, req.Value); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}

func (r *rpcHDFSServer) Delete(ctx context.Context, key *fspb.Key) (*fspb.Error, error) {
	if err := r.fs.Delete(key.Key); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}

func (r *rpcHDFSServer) ListPeer(ctx context.Context, empty *emptypb.Empty) (*fspb.PeerList, error) {
	list := r.fs.Peer().PList()
	pbList := make([]*fspb.PeerInfo, 0, len(list))
	for _, v := range list {
		pbList = append(pbList, &fspb.PeerInfo{
			Name:   v.PName(),
			Addr:   v.PAddr(),
			Stat:   int64(v.PStat()),
			Action: int64(peers.P_ACTION_NONE),
		})
	}
	return &fspb.PeerList{
		Peers: pbList,
	}, nil
}

func (r *rpcHDFSServer) PeerSync(ctx context.Context, pi *fspb.PeerInfo) (*fspb.Error, error) {
	if err := r.fs.Peer().PSync(DPeerInfo{
		PeerName: pi.Name,
		PeerAddr: pi.Addr,
		PeerStat: peers.PeerStatType(pi.Stat),
	}, peers.PeerActionType(pi.GetAction())); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}

func (r *rpcHDFSServer) run(port string) {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return
	}
	fspb.RegisterPeerServiceServer(r.s, r)
	fspb.RegisterHashFileSystemServiceServer(r.s, r)
	err = r.s.Serve(l)
	if err != nil {
		log.Println("[RPC Server] Server shutdown:", err)
		return
	}
}
