package dfs

import (
	"time"

	"github.com/ciiim/cloudborad/internal/dfs/peers"
	"github.com/ciiim/cloudborad/internal/random"
)

var (
	DefaultSyncSettings = SyncSettings{
		syncInterval: time.Second,
		syncTimeout:  5 * time.Second,
		syncPickNum:  3,
	}
)

type SyncSettings struct {

	// 同步间隔
	syncInterval time.Duration

	// 最晚同步时间
	syncTimeout time.Duration

	// 同步节点数量
	syncPickNum int
}

// pingResponseObject
type pingRO struct {
	needSync bool
	version  int64
}

type dgossip struct {
	peerinfos []struct {
		id   int64
		name string
		addr string
	}
	version int64
}

type dgossipResponse struct {
	version int64
	boring  bool
}

func (p *DPeer) pickPeers() (pickedList []peers.PeerInfo) {
	list := p.hashMap.List()
	pickedNum := 0
	// 如果节点数量小于同步节点数量，直接返回
	if len(list) <= p.syncSettings.syncPickNum {
		return list
	}
	randList := random.Number[int](len(list))

	// 如果节点上次同步时间超过最晚同步时间，选取该节点
	for i, v := range list {
		if pickedNum >= p.syncSettings.syncPickNum {
			return
		}
		dv := v.(*DPeerInfo)
		if dv.LastPingTime.Add(p.syncSettings.syncTimeout).After(time.Now()) {
			pickedNum++
			pickedList = append(pickedList, dv)
			resetPingTime(dv)
			randList.Remove(i)
		}
	}
	//若节点数量不足，随机选取节点
	for pickedNum < p.syncSettings.syncPickNum {
		pickedNum++
		pickedList = append(pickedList, list[randList.Get()])
	}
	return pickedList
}

func resetPingTime(pi *DPeerInfo) {
	pi.LastPingTime = time.Now()
}

type pingResponse struct {
	needSync bool
	target   DPeerInfo
}

func (p *DPeer) ping() pingResponse {
	return pingResponse{}
}

func (p *DPeer) handlePing(pr pingResponse) {
	if pr.needSync {
		p.pull(pr.target)
	}
}

func (p *DPeer) pong() error {
	return nil
}

// 请求拉取对方所有节点信息
func (p *DPeer) pull(target DPeerInfo) (dgossip, error) {

}

func (p *DPeer) gossip(msg dgossip, dpi DPeerInfo) (dgossipResponse, error) {

}
