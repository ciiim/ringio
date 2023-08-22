package fs

import (
	"context"
	"log"
	"strings"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

type DPeer struct {
	info    DPeerInfo
	hashMap *peers.CMap
}

var _ peers.Peer = (*DPeer)(nil)

type DPeerInfo struct {
	PeerName string             `json:"peer_name"`
	PeerAddr string             `json:"peer_addr"` //include port e.g. 10.10.1.5:9631
	PeerStat peers.PeerStatType `json:"peer_stat"`
}

func NewDPeerInfo(name, addr string) DPeerInfo {
	return DPeerInfo{
		PeerName: name,
		PeerAddr: addr,
		PeerStat: peers.P_STAT_ONLINE,
	}
}

var _ peers.PeerInfo = (*DPeerInfo)(nil)

func NewDPeer(name, addr string, replicas int, peersHashFn peers.CHash) *DPeer {
	dlog.debug("NewDPeer", "name: %s, addr: %s", name, addr)
	info := DPeerInfo{
		PeerName: name,
		PeerAddr: addr,
		PeerStat: peers.P_STAT_ONLINE,
	}
	p := &DPeer{
		info:    info,
		hashMap: peers.NewCMap(replicas, peersHashFn),
	}
	p.hashMap.Add(info)
	return p
}

func (p DPeer) Get(pi peers.PeerInfo, key string) peers.PeerResult {
	client := newRpcClient(p.info.Port())
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	file, err := client.get(ctx, pi, key)
	if err != nil {
		return peers.PeerResult{
			Err: err,
		}
	}
	return peers.PeerResult{
		Err:  nil,
		Data: file.Data(),
		Info: file.Stat(),
		Pi:   file.Stat().PeerInfo(),
	}
}

func (p DPeer) Put(pi peers.PeerInfo, key string, filename string, value []byte) peers.PeerResult {
	res := peers.PeerResult{}
	client := newRpcClient(p.info.Port())

	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	res.Err = client.put(ctx, pi, key, filename, value)
	return res
}

func (p DPeer) Delete(pi peers.PeerInfo, key string) peers.PeerResult {
	res := peers.PeerResult{}
	client := newRpcClient(p.info.Port())

	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	res.Err = client.delete(ctx, pi, key)
	return res
}

func (p DPeer) PName() string {
	return p.info.PeerName
}

func (p DPeer) PAddr() string {
	return p.info.PeerAddr
}

func (p DPeer) Pick(key string) peers.PeerInfo {
	return p.hashMap.Get(key)
}

func (p DPeer) PAdd(pis ...peers.PeerInfo) {
	p.hashMap.Add(pis...)
}

func (p DPeer) PDel(pis ...peers.PeerInfo) {
	p.hashMap.Del(pis...)
}

/*
recieve peer action from other peer
source peer - pi_in
*/
func (p DPeer) PSync(pi_in peers.PeerInfo, action peers.PeerActionType) error {
	dlog.debug("PSync", "pi_in: %v, action: %s", pi_in, action.String())
	if pi_in.Equal(p.info) {
		log.Println("[Peer] Cannot Operate myself")
		return nil
	}
	var err error
	switch action {
	case peers.P_ACTION_JOIN:
		// notify other peers - action P_ACTION_NEW
		client := newRpcClient(pi_in.Port())
		ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
		defer cancel()
		list := p.PList()
		err = client.peerActionTo(ctx, pi_in, peers.P_ACTION_NEW, list...)
	case peers.P_ACTION_QUIT:
		// remove peer from hashMap
		p.PDel(pi_in)
	case peers.P_ACTION_NEW:
		// add peer to hashMap
		p.PAdd(pi_in)
	}
	return err
}

/*
send peer action to other peer

pi_to - destination peer
*/
func (p DPeer) PActionTo(action peers.PeerActionType, pi_to ...peers.PeerInfo) error {
	dlog.debug("PActionTo", "action: %s, pi_to: %v", action.String(), pi_to)
	client := newRpcClient(p.info.Port())
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return client.peerActionTo(ctx, p.info, action, pi_to...)
}

func (p DPeer) GetPeerListFromPeer(pi peers.PeerInfo) ([]peers.PeerInfo, error) {
	client := newRpcClient(p.info.Port())
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	list, err := client.getPeerList(ctx, pi)
	if err != nil {
		return nil, err
	}
	peerList := make([]peers.PeerInfo, 0, len(list))
	for _, v := range list {
		peerList = append(peerList, NewDPeerInfo(v.PName(), v.PAddr()))
	}
	return peerList, nil
}

func (p DPeer) PNext(key string) peers.PeerInfo {
	return p.hashMap.GetPeerNext(key, 1)
}

func (p DPeer) PList() []peers.PeerInfo {
	return p.hashMap.List()
}

func (p DPeer) Info() peers.PeerInfo {
	return p.info
}

func (pi DPeerInfo) Equal(other peers.PeerInfo) bool {
	o := other.(DPeerInfo)
	return pi.PeerName == o.PeerName && pi.PeerAddr == o.PeerAddr
}

func (pi DPeerInfo) PName() string {
	return pi.PeerName
}

func (pi DPeerInfo) PAddr() string {
	return pi.PeerAddr
}

func (pi DPeerInfo) PStat() peers.PeerStatType {
	return pi.PeerStat
}

func (pi DPeerInfo) Port() string {
	t := strings.Split(pi.PeerAddr, ":")
	if len(t) != 2 {
		return ""
	}
	return t[len(t)-1]
}
