package ringio

import (
	"context"

	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/types"
)

func (r *rpcServer) MakeDir(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.tfs.MakeDir(req.Space, req.Base, req.Name)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcServer) RenameDir(ctx context.Context, req *fspb.RenameDirRequest) (*fspb.Error, error) {
	err := r.tfs.RenameDir(req.Src.Space, req.Src.Base, req.Src.Name, req.NewName)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcServer) DeleteDir(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.tfs.DeleteDir(req.Space, req.Base, req.Name)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcServer) GetDirSub(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Subs, error) {
	subs, err := r.tfs.GetDirSub(req.Space, req.Base, req.Name)
	return &fspb.Subs{SubInfo: SubsToPbSubs(subs)}, err
}

func (r *rpcServer) NewSpace(ctx context.Context, space *fspb.NewSpaceRequest) (*fspb.Error, error) {
	err := r.tfs.Local().NewLocalSpace(space.Space, types.Byte(space.Cap))
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcServer) DeleteSpace(ctx context.Context, space *fspb.SpaceRequest) (*fspb.Error, error) {
	err := r.tfs.Local().DeleteLocalSpace(space.Space)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcServer) GetMetadata(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.BytesData, error) {
	data, err := r.tfs.GetMetadata(req.Space, req.Base, req.Name)
	if err != nil {
		return nil, err
	}
	return &fspb.BytesData{Data: data}, nil
}

func (r *rpcServer) PutMetadata(ctx context.Context, req *fspb.PutMetadataRequest) (*fspb.Error, error) {
	err := r.tfs.PutMetadata(req.Src.Space, req.Src.Base, req.Src.Name, req.Src.Hash, req.Metadata)
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}

func (r *rpcServer) DeleteMetadata(ctx context.Context, req *fspb.TreeFileSystemBasicRequest) (*fspb.Error, error) {
	err := r.tfs.DeleteMetadata(req.GetSpace(), req.GetBase(), req.GetName())
	if err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}
