package fs

import (
	"context"
	"log"
	"net"

	"github.com/ciiim/cloudborad/internal/fs/fspb"
	"github.com/ciiim/cloudborad/internal/fs/peers"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type rpcTDFSServer struct {
	s *grpc.Server

	fs TreeDFileSystemI
	fspb.UnimplementedPeerServiceServer
	fspb.UnimplementedTreeFileSystemServiceServer
}

func newRPCTDFSServer(fs TreeDFileSystemI) *rpcTDFSServer {
	return &rpcTDFSServer{
		fs: fs,
		s:  grpc.NewServer(),
	}
}

func (r *rpcTDFSServer) MakeDir(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.fs.MakeDir(req.Space, req.Base, req.Name)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcTDFSServer) RenameDir(ctx context.Context, req *fspb.RenameDirRequest) (*fspb.Error, error) {
	err := r.fs.RenameDir(req.Src.Space, req.Src.Base, req.Src.Name, req.NewName)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcTDFSServer) DeleteDir(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.fs.DeleteDir(req.Space, req.Base, req.Name)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcTDFSServer) GetDirSub(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Subs, error) {
	subs, err := r.fs.GetDirSub(req.Space, req.Base, req.Name)
	return &fspb.Subs{SubInfo: subsToPbSubs(subs)}, err
}

func (r *rpcTDFSServer) NewSpace(ctx context.Context, space *fspb.NewSpaceRequest) (*fspb.Error, error) {
	err := r.fs.NewSpace(space.Space, Byte(space.Cap))
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcTDFSServer) DeleteSpace(ctx context.Context, space *fspb.SpaceRequest) (*fspb.Error, error) {
	err := r.fs.DeleteSpace(space.Space)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcTDFSServer) GetMetadata(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.BytesData, error) {
	data, err := r.fs.GetMetadata(req.Space, req.Base, req.Name)
	if err != nil {
		return nil, err
	}
	return &fspb.BytesData{Data: data}, nil
}

func (r *rpcTDFSServer) PutMetadata(ctx context.Context, req *fspb.PutMetadataRequest) (*fspb.PutMetadataResponse, error) {
	path, err := r.fs.PutMetadata(req.Src.Space, req.Src.Base, req.Src.Name, req.Src.Hash, req.Metadata)
	if err != nil {
		return &fspb.PutMetadataResponse{
			Path: "",
			Err:  &fspb.Error{Err: err.Error()},
		}, nil
	}
	return &fspb.PutMetadataResponse{
		Path: path,
	}, nil
}

func (r *rpcTDFSServer) DeleteMetadata(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.fs.DeleteMetadata(req.Space, req.Base, req.Name)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcTDFSServer) ListPeer(ctx context.Context, empty *emptypb.Empty) (*fspb.PeerList, error) {
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

func (r *rpcTDFSServer) PeerSync(ctx context.Context, pi *fspb.PeerInfo) (*fspb.Error, error) {
	if err := r.fs.Peer().PSync(DPeerInfo{
		PeerName: pi.Name,
		PeerAddr: pi.Addr,
		PeerStat: peers.PeerStatType(pi.Stat),
	}, peers.PeerActionType(pi.GetAction())); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}

func (r *rpcTDFSServer) Serve(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fspb.RegisterPeerServiceServer(r.s, r)
	fspb.RegisterTreeFileSystemServiceServer(r.s, r)
	if err := r.s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
