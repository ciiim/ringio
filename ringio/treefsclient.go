package ringio

import (
	"context"
	"errors"

	"github.com/ciiim/cloudborad/chunkpool"
	dlogger "github.com/ciiim/cloudborad/debug"
	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/tree"
	"github.com/ciiim/cloudborad/storage/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type rpcTreeClient struct {
}

func newRPCHashClient(pool *chunkpool.ChunkPool) *rpcHashClient {
	return &rpcHashClient{
		defaultBufferSize: DefaultBufferSize,
		pool:              pool,
	}
}

func newRPCTreeClient() *rpcTreeClient {
	return &rpcTreeClient{}
}

func (r *rpcTreeClient) getMetadata(ctx context.Context, ni *node.Node, space string, base string, name string) ([]byte, error) {
	dlogger.Dlog.LogDebugf("[RPC Client]", "GetMetadata from %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &fspb.TreeFileSystemBasicRequest{
		Space: space,
		Base:  base,
		Name:  name,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (r *rpcTreeClient) putMetadata(ctx context.Context, ni *node.Node, space string, base string, name string, data []byte) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "PutMetadata to %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	resp, err := client.PutMetadata(ctx, &fspb.PutMetadataRequest{
		Src: &fspb.TreeFileSystemBasicRequest{
			Space: space,
			Base:  base,
			Name:  name,
		},
		Metadata: data,
	})
	respErr := errors.New(resp.Err)
	if respErr != nil {
		return err
	}

	return respErr
}

func (r *rpcTreeClient) deleteMetadata(ctx context.Context, ni *node.Node, space string, base string, name string, hash []byte) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "DeleteMetadata in %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	_, err = client.DeleteMetadata(ctx, &fspb.TreeFileSystemBasicRequest{
		Space: space,
		Base:  base,
		Name:  name,
		Hash:  hash,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *rpcTreeClient) makeDir(ctx context.Context, ni *node.Node, space string, base string, dir string) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "MakeDir in %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	_, err = client.MakeDir(ctx, &fspb.TreeFileSystemBasicRequest{
		Space: space,
		Base:  base,
		Name:  dir,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *rpcTreeClient) renameDir(ctx context.Context, ni *node.Node, space string, base string, dir string, newName string) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "RenameDir in %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	_, err = client.RenameDir(ctx, &fspb.RenameDirRequest{
		Src: &fspb.TreeFileSystemBasicRequest{
			Space: space,
			Base:  base,
			Name:  dir,
		},
		NewName: newName,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *rpcTreeClient) deleteDir(ctx context.Context, ni *node.Node, space string, base string, dir string) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "DeleteDir in %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	_, err = client.DeleteDir(ctx, &fspb.TreeFileSystemBasicRequest{
		Space: space,
		Base:  base,
		Name:  dir,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *rpcTreeClient) getDirSub(ctx context.Context, ni *node.Node, space string, base string, dir string) ([]*tree.SubInfo, error) {
	dlogger.Dlog.LogDebugf("[RPC Client]", "GetDirSub from %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	resp, err := client.GetDirSub(ctx, &fspb.TreeFileSystemBasicRequest{
		Space: space,
		Base:  base,
		Name:  dir,
	})
	if err != nil {
		return nil, err
	}
	return PbSubsToSubs(resp.SubInfo), nil
}

func (r *rpcTreeClient) newSpace(ctx context.Context, ni *node.Node, space string, cap types.Byte) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "NewSpace in %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	_, err = client.NewSpace(ctx, &fspb.NewSpaceRequest{
		Space: space,
		Cap:   int64(cap),
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *rpcTreeClient) deleteSpace(ctx context.Context, ni *node.Node, space string) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "DeleteSpace in %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewTreeFileSystemServiceClient(conn)
	_, err = client.DeleteSpace(ctx, &fspb.SpaceRequest{
		Space: space,
	})
	if err != nil {
		return err
	}
	return nil
}
