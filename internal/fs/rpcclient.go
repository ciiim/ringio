package fs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ciiim/cloudborad/internal/fs/peers"

	"github.com/ciiim/cloudborad/internal/fs/fspb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	RPC_TDFS_PORT = "9631"
	RPC_HDFS_PORT = "9632"
	_RPC_TIMEOUT  = time.Second * 5
)

type rpcHashClient struct {
}

type rpcTreeClient struct {
}

type rpcPeerClient struct {
}

func newRPCHashClient() *rpcHashClient {
	return &rpcHashClient{}
}

func newRPCPeerClient() *rpcPeerClient {
	return &rpcPeerClient{}
}

func newRPCTreeClient() *rpcTreeClient {
	return &rpcTreeClient{}
}

func ctxWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), _RPC_TIMEOUT)
}

/*

Peer Method Start

*/

func (c *rpcPeerClient) peerActionTo(ctx context.Context, target peers.PeerInfo, action peers.PeerActionType, pis ...peers.PeerInfo) error {
	for _, pi := range pis {
		log.Printf("[RPC Client] PeerAction: %s to %s\n", action.String(), pi.PAddr())
		conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("[RPC Client] Dial %s error: %s", pi.PAddr(), err.Error())
			continue
		}

		client := fspb.NewPeerServiceClient(conn)
		_, err = client.PeerSync(ctx, &fspb.PeerInfo{
			Name:   target.PName(),
			Addr:   target.PAddr(),
			Stat:   int64(target.PStat()),
			Action: int64(action),
		})
		conn.Close()
		if err != nil {
			log.Printf("[RPC Client] PeerAction %s to %s error: %s", action.String(), pi.PAddr(), err.Error())
			continue
		}
	}
	return nil
}

func (c *rpcPeerClient) getPeerList(ctx context.Context, pi peers.PeerInfo) ([]peers.PeerInfo, error) {
	log.Printf("[RPC Client] GetPeerList from %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := fspb.NewPeerServiceClient(conn)
	resp, err := client.ListPeer(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	var pis []peers.PeerInfo
	for _, pi := range resp.Peers {
		pis = append(pis, NewDPeerInfo(pi.Name, pi.Addr))
	}
	return pis, nil
}

/*

Peer Method End

*/

/*

rpcHashClient Method Start

*/

func (c *rpcHashClient) get(ctx context.Context, pi peers.PeerInfo, key string) (HashDFile, error) {
	log.Printf("[RPC Client] Get from %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return HashDFile{}, err
	}
	defer conn.Close()

	client := fspb.NewHashFileSystemServiceClient(conn)
	resp, err := client.Get(ctx, &fspb.Key{Key: key})
	if err != nil {
		return HashDFile{}, err
	}
	hfi := pBFileInfoToHashFileInfo(resp.FileInfo)
	return HashDFile{
		data: resp.Data,
		info: HashDFileInfo{
			HashFileInfo: hfi,
			DPeerInfo: DPeerInfo{
				PeerName: resp.PeerInfo.Name,
				PeerAddr: resp.PeerInfo.Addr,
				PeerStat: peers.PeerStatType(resp.PeerInfo.Stat),
			},
		},
	}, nil
}

func (c *rpcHashClient) put(ctx context.Context, pi peers.PeerInfo, key string, filename string, value []byte) error {
	log.Printf("[RPC Client] Put to %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewHashFileSystemServiceClient(conn)

	_, err = client.Put(ctx, &fspb.PutRequest{Key: &fspb.Key{Key: key}, Filename: filename, Value: value})
	if err != nil {
		return err
	}
	return nil
}

func (c *rpcHashClient) delete(ctx context.Context, pi peers.PeerInfo, key string) error {
	log.Printf("[RPC Client] Delete file in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewHashFileSystemServiceClient(conn)
	_, err = client.Delete(ctx, &fspb.Key{Key: key})
	if err != nil {
		return err
	}
	return nil
}

/*

rpcHashClient Method End

*/

/*

newRPCTreeClient Method Start

*/

func (r *rpcTreeClient) getMetadata(ctx context.Context, pi peers.PeerInfo, space string, base string, name string) ([]byte, error) {
	log.Printf("[RPC Client] GetMetadata from %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (r *rpcTreeClient) putMetadata(ctx context.Context, pi peers.PeerInfo, space string, base string, name string, data []byte) error {
	log.Printf("[RPC Client] PutMetadata to %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (r *rpcTreeClient) deleteMetadata(ctx context.Context, pi peers.PeerInfo, space string, base string, name, hash string) error {
	log.Printf("[RPC Client] DeleteMetadata in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (r *rpcTreeClient) makeDir(ctx context.Context, pi peers.PeerInfo, space string, base string, dir string) error {
	log.Printf("[RPC Client] MakeDir in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (r *rpcTreeClient) renameDir(ctx context.Context, pi peers.PeerInfo, space string, base string, dir string, newName string) error {
	log.Printf("[RPC Client] RenameDir in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (r *rpcTreeClient) deleteDir(ctx context.Context, pi peers.PeerInfo, space string, base string, dir string) error {
	log.Printf("[RPC Client] DeleteDir in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (r *rpcTreeClient) getDirSub(ctx context.Context, pi peers.PeerInfo, space string, base string, dir string) ([]SubInfo, error) {
	log.Printf("[RPC Client] GetDirSub from %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	return pbSubsToSubs(resp.SubInfo), nil
}

func (r *rpcTreeClient) newSpace(ctx context.Context, pi peers.PeerInfo, space string, cap Byte) error {
	log.Printf("[RPC Client] NewSpace in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (r *rpcTreeClient) deleteSpace(ctx context.Context, pi peers.PeerInfo, space string) error {
	log.Printf("[RPC Client] DeleteSpace in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

/*

newRPCTreeClient Method End

*/
