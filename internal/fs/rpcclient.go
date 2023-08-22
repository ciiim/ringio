package fs

import (
	"context"
	"log"
	"time"

	"github.com/ciiim/cloudborad/internal/fs/peers"

	"github.com/ciiim/cloudborad/internal/fs/fspb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	FRONT_PORT      = "9631"
	FILE_STORE_PORT = "9632"
	_RPC_TIMEOUT    = time.Second * 5
)

type rpcClient struct {
	port string
}

func newRpcClient(port string) *rpcClient {
	return &rpcClient{
		port: port,
	}
}

func (c *rpcClient) get(ctx context.Context, pi peers.PeerInfo, key string) (File, error) {
	log.Printf("[RPC Client] Get from %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr()+":"+c.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := fspb.NewPeerServiceClient(conn)
	resp, err := client.Get(ctx, &fspb.Key{Key: key})
	if err != nil {
		return nil, err
	}

	if resp.FileInfo.IsDir {
		tfi := pbFileInfoToTreeFileInfo(resp.FileInfo)
		return DTreeFile{
			data: resp.Data,
			info: DTreeFileInfo{
				TreeFileInfo: tfi,
				DPeerInfo: DPeerInfo{
					PeerName: resp.PeerInfo.Name,
					PeerAddr: resp.PeerInfo.Addr,
					PeerStat: peers.PeerStatType(resp.PeerInfo.Stat),
				},
			},
		}, nil
	} else {
		bfi := pBFileInfoToBasicFileInfo(resp.FileInfo)
		return DistributeFile{
			data: resp.Data,
			info: DistributeFileInfo{
				BasicFileInfo: bfi,
				DPeerInfo: DPeerInfo{
					PeerName: resp.PeerInfo.Name,
					PeerAddr: resp.PeerInfo.Addr,
					PeerStat: peers.PeerStatType(resp.PeerInfo.Stat),
				},
			},
		}, nil
	}

}

func (c *rpcClient) put(ctx context.Context, pi peers.PeerInfo, key string, filename string, value []byte) error {
	log.Printf("[RPC Client] Put to %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr()+":"+c.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewPeerServiceClient(conn)

	_, err = client.Put(ctx, &fspb.PutRequest{Key: &fspb.Key{Key: key}, Filename: filename, Value: value})
	if err != nil {
		return err
	}
	return nil
}

func (c *rpcClient) delete(ctx context.Context, pi peers.PeerInfo, key string) error {
	log.Printf("[RPC Client] Delete file in %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr()+":"+c.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewPeerServiceClient(conn)
	_, err = client.Delete(ctx, &fspb.Key{Key: key})
	if err != nil {
		return err
	}
	return nil
}

func (c *rpcClient) peerActionTo(ctx context.Context, target peers.PeerInfo, action peers.PeerActionType, pis ...peers.PeerInfo) error {
	for _, pi := range pis {
		log.Printf("[RPC Client] PeerAction: %d to %s\n", action, pi.PAddr())
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

func (c *rpcClient) getPeerList(ctx context.Context, pi peers.PeerInfo) ([]peers.PeerInfo, error) {
	log.Printf("[RPC Client] GetPeerList from %s", pi.PAddr())
	conn, err := grpc.Dial(pi.PAddr()+":"+c.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
