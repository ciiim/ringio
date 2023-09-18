package remote

import (
	"context"

	"github.com/ciiim/cloudborad/internal/fs/fspb"
	"github.com/ciiim/cloudborad/internal/fs/peers"
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

func (r *rpcFSServer) PeerSync(ctx context.Context, pi *fspb.PeerInfo) (*fspb.Error, error) {
	if err := r.peerService.PSync(DPeerInfo{
		PeerName: pi.Name,
		PeerAddr: DAddr(pi.Addr),
		PeerStat: peers.PeerStatType(pi.Stat),
	}, peers.PeerActionType(pi.GetAction())); err != nil {
		return &fspb.Error{Err: err.Error()}, err
	}
	return &fspb.Error{}, nil
}
