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

// RPC Server Port : 9632

type rpcServer struct {
	fs DistributeFileSystem
	fspb.UnimplementedPeerServiceServer
}

func newRpcServer(fs DistributeFileSystem) *rpcServer {
	return &rpcServer{
		fs: fs,
	}
}

func (r *rpcServer) Get(ctx context.Context, key *fspb.Key) (*fspb.GetResponse, error) {
	file, err := r.fs.Get(key.Key)
	if err != nil {
		return nil, nil
	}
	fi := file.Stat()

	//convert subdir to pb.DirInfo
	var pbSubDir []*fspb.DirInfo
	if fi.IsDir() {
		pbSubDir = make([]*fspb.DirInfo, 0, len(fi.SubDir()))
		for _, v := range fi.SubDir() {
			pbSubDir = append(pbSubDir, &fspb.DirInfo{
				DirName: v.Name(),
			})
		}
	}

	return &fspb.GetResponse{
		Data: file.Data(),
		FileInfo: &fspb.FileInfo{
			FileName: fi.Name(),
			BasePath: fi.Path(),
			Hash:     fi.Hash(),
			Size:     fi.Size(),
			IsDir:    fi.IsDir(),
			DirInfo:  pbSubDir,
		},
		PeerInfo: &fspb.PeerInfo{
			Name:   fi.PeerInfo().PName(),
			Addr:   fi.PeerInfo().PAddr(),
			Stat:   int64(fi.PeerInfo().PStat()),
			Action: int64(peers.P_ACTION_NONE),
		},
	}, nil
}

func (r *rpcServer) Put(ctx context.Context, req *fspb.PutRequest) (*emptypb.Empty, error) {
	if err := r.fs.Store(req.Key.Key, req.Filename, req.Value); err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (r *rpcServer) Delete(ctx context.Context, key *fspb.Key) (*emptypb.Empty, error) {
	if err := r.fs.Delete(key.Key); err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (r *rpcServer) ListPeer(ctx context.Context, empty *emptypb.Empty) (*fspb.PeerList, error) {
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

func (r *rpcServer) GetPeerAction(ctx context.Context, pi *fspb.PeerInfo) (*emptypb.Empty, error) {
	if err := r.fs.Peer().PSync(DPeerInfo{
		name: pi.Name,
		addr: pi.Addr,
		stat: peers.PeerStatType(pi.Stat),
	}, peers.PeerActionType(pi.GetAction())); err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (r *rpcServer) run(port string) {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return
	}
	log.Printf("[RPC Server] Listen: %s\n", l.Addr())
	s := grpc.NewServer()
	fspb.RegisterPeerServiceServer(s, r)
	err = s.Serve(l)
	if err != nil {
		log.Println("[RPC Server] Server shutdown:", err)
		return
	}
}
