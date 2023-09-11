package fs

import (
	"context"
	"log"
	"strings"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

type DAddr string

func (a DAddr) String() string {
	return string(a)
}

func (a DAddr) Port() string {
	t := strings.Split(string(a), ":")
	if len(t) != 2 {
		return ""
	}
	return t[len(t)-1]
}

func (a DAddr) IP() string {
	t := strings.Split(string(a), ":")
	if len(t) != 2 {
		return ""
	}
	return t[0]
}

type DPeer struct {
	info    DPeerInfo
	hashMap *peers.CMap
}

var _ peers.Peer = (*DPeer)(nil)

type DPeerInfo struct {
	PeerID   int64              `json:"peer_id"`
	PeerName string             `json:"peer_name"`
	PeerAddr peers.Addr         `json:"peer_addr"` //include port e.g. 10.10.1.5:9631
	PeerStat peers.PeerStatType `json:"peer_stat"`
}

func NewDPeerInfo(name, addr string) DPeerInfo {
	return DPeerInfo{
		PeerName: name,
		PeerAddr: DAddr(addr),
		PeerStat: peers.P_STAT_ONLINE,
	}
}

var _ peers.PeerInfo = (*DPeerInfo)(nil)

func NewDPeer(name, addr string, replicas int, peersHashFn peers.CHash) *DPeer {
	dlog.debug("NewDPeer", "name: %s, addr: %s", name, addr)
	info := DPeerInfo{
		PeerName: name,
		PeerAddr: DAddr(addr),
		PeerStat: peers.P_STAT_ONLINE,
	}
	p := &DPeer{
		info:    info,
		hashMap: peers.NewCMap(replicas, peersHashFn),
	}
	p.hashMap.Add(info)
	return p
}

func (p DPeer) PName() string {
	return p.info.PeerName
}

func (p DPeer) PAddr() peers.Addr {
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
		client := newRPCPeerClient()
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
	client := newRPCPeerClient()
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	return client.peerActionTo(ctx, p.info, action, pi_to...)
}

func (p DPeer) GetPeerListFromPeer(pi peers.PeerInfo) ([]peers.PeerInfo, error) {
	client := newRPCPeerClient()
	ctx, cancel := context.WithTimeout(context.Background(), _RPC_TIMEOUT)
	defer cancel()
	list, err := client.getPeerList(ctx, pi)
	if err != nil {
		return nil, err
	}
	peerList := make([]peers.PeerInfo, 0, len(list))
	for _, v := range list {
		peerList = append(peerList, NewDPeerInfo(v.PName(), v.PAddr().String()))
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

func (pi DPeerInfo) PAddr() peers.Addr {
	return pi.PeerAddr
}

func (pi DPeerInfo) PStat() peers.PeerStatType {
	return pi.PeerStat
}

func PeerInfoListToDpeerInfoList(list []peers.PeerInfo) []DPeerInfo {
	dpeerList := make([]DPeerInfo, 0, len(list))
	for _, v := range list {
		dpeerList = append(dpeerList, v.(DPeerInfo))
	}
	return dpeerList
}
