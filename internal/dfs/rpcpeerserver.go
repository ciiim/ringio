package dfs

import (
	"context"

	"github.com/ciiim/cloudborad/internal/dfs/fspb"
	"github.com/ciiim/cloudborad/internal/dfs/peers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (r *rpcFSServer) ListPeer(ctx context.Context, empty *emptypb.Empty) (*fspb.PeerList, error) {
	list := r.peerService.PList()
	pbList := make([]*fspb.PeerInfo, 0, len(list))
	for _, v := range list {
		pbList = append(pbList, &fspb.PeerInfo{
			Name:   v.PName(),
			Addr:   v.PAddr().String(),
			Stat:   int64(v.PStat()),
			Action: int64(peers.P_ACTION_NONE),
		})
	}
	return &fspb.PeerList{
		Peers: pbList,
	}, nil
}

/*
Deprecated: use Sync instead
*/
func (r *rpcFSServer) PeerSync(ctx context.Context, pi *fspb.PeerInfo) (*fspb.Error, error) {
	if err := r.peerService.PHandleSyncAction(DPeerInfo{
		PeerName: pi.Name,
		PeerAddr: DAddr(pi.Addr),
		PeerStat: peers.PeerStatType(pi.Stat),
	}, peers.PeerActionType(pi.GetAction())); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}

func (r *rpcFSServer) Sync(ctx context.Context, ping *fspb.SyncPing) (*fspb.SyncPong, error) {
	localVersion := r.peerService.Info().PVersion()
	remoteVersion := ping.Version
	if localVersion <= remoteVersion {
		return &fspb.SyncPong{
			NeedSync: false,
			Version:  localVersion,
		}, nil
	}
	return &fspb.SyncPong{
		NeedSync: true,
		Version:  localVersion,
	}, nil
}

func (r *rpcFSServer) SyncPull(ctx context.Context, pullReq *fspb.SyncPullRequest) (*fspb.SyncPullResponse, error) {
	peerList := r.peerService.PList()
	pbList := make([]*fspb.PeerInfo, 0, len(peerList))
	for _, v := range peerList {
		pbList = append(pbList, &fspb.PeerInfo{
			Name:   v.PName(),
			Addr:   v.PAddr().String(),
			Stat:   int64(v.PStat()),
			Action: int64(peers.P_ACTION_NONE),
		})
	}
	return &fspb.SyncPullResponse{
		Peers:   pbList,
		Version: r.peerService.Info().PVersion(),
	}, nil
}

func (r *rpcFSServer) Gossip(ctx context.Context, gossipMsg *fspb.GossipMsg) (*fspb.GossipAck, error) {

}
