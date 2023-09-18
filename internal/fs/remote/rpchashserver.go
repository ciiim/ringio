package remote

import (
	"context"

	"github.com/ciiim/cloudborad/internal/fs/peers"

	"github.com/ciiim/cloudborad/internal/fs/fspb"
)

func (r *rpcFSServer) Get(ctx context.Context, key *fspb.Key) (*fspb.GetResponse, error) {
	file, err := r.hfs.Get(key.Key)
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
			Addr:   fi.PeerInfo().PAddr().String(),
			Stat:   int64(fi.PeerInfo().PStat()),
			Action: int64(peers.P_ACTION_NONE),
		},
	}, nil
}

func (r *rpcFSServer) Put(ctx context.Context, req *fspb.PutRequest) (*fspb.Error, error) {
	if err := r.hfs.Store(req.Key.Key, req.Filename, req.Value); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) Delete(ctx context.Context, key *fspb.Key) (*fspb.Error, error) {
	if err := r.hfs.Delete(key.Key); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}
