package fs

import (
	"context"

	"github.com/ciiim/cloudborad/internal/fs/fspb"
)

func (r *rpcFSServer) MakeDir(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.tfs.MakeDir(req.Space, req.Base, req.Name)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) RenameDir(ctx context.Context, req *fspb.RenameDirRequest) (*fspb.Error, error) {
	err := r.tfs.RenameDir(req.Src.Space, req.Src.Base, req.Src.Name, req.NewName)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) DeleteDir(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.tfs.DeleteDir(req.Space, req.Base, req.Name)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) GetDirSub(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Subs, error) {
	subs, err := r.tfs.GetDirSub(req.Space, req.Base, req.Name)
	return &fspb.Subs{SubInfo: subsToPbSubs(subs)}, err
}

func (r *rpcFSServer) NewSpace(ctx context.Context, space *fspb.NewSpaceRequest) (*fspb.Error, error) {
	err := r.tfs.NewSpace(space.Space, Byte(space.Cap))
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) DeleteSpace(ctx context.Context, space *fspb.SpaceRequest) (*fspb.Error, error) {
	err := r.tfs.DeleteSpace(space.Space)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) GetMetadata(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.BytesData, error) {
	data, err := r.tfs.GetMetadata(req.Space, req.Base, req.Name)
	if err != nil {
		return nil, err
	}
	return &fspb.BytesData{Data: data}, nil
}

func (r *rpcFSServer) PutMetadata(ctx context.Context, req *fspb.PutMetadataRequest) (*fspb.Error, error) {
	err := r.tfs.PutMetadata(req.Src.Space, req.Src.Base, req.Src.Name, req.Src.Hash, req.Metadata)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) DeleteMetadata(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.tfs.DeleteMetadata(req.Space, req.Base, req.Name, req.Hash)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}
