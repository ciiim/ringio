// about peers
package peers

import "errors"

var (
	ErrPeerNotFound = errors.New("peer not found")
	PeerLocal       = &LocalPeer{}
)

type PeerStatType int
type PeerActionType int

const (
	P_STAT_ONLINE PeerStatType = iota
	P_STAT_OFFLINE
	P_STAT_REMOVED
)
const (
	P_ACTION_NONE PeerActionType = iota

	// 新节点接入集群
	P_ACTION_JOIN

	// 通知集群中其余节点有新节点加入
	P_ACTION_NEW

	// 节点退出集群 (主动退出)
	P_ACTION_QUIT
)

type Peer interface {
	PName() string
	PAddr() string
	Pick(key string) PeerInfo
	Info() PeerInfo
	PeerGetSetDeleter
	PeerOperator
}

type PeerInfoList []PeerInfo

type PeerInfo interface {
	Equal(pi PeerInfo) bool
	PName() string
	PAddr() string
	Port() string
	PStat() PeerStatType
}

type PeerResult struct {
	Err  error
	Data []byte
	Pi   PeerInfo
	Info any
}

type PeerGetSetDeleter interface {
	Get(pi PeerInfo, key string) PeerResult
	Put(pi PeerInfo, key string, filename string, value []byte) PeerResult
	Delete(pi PeerInfo, key string) PeerResult
}

type PeerOperator interface {
	PAdd(pis ...PeerInfo)
	PDel(pis ...PeerInfo)
	PNext(key string) PeerInfo
	PSync(pi PeerInfo, action PeerActionType) error
	PActionTo(action PeerActionType, pi_to ...PeerInfo) error
	PList() PeerInfoList
}

type LocalPeer struct {
}

type LocalPeerInfo struct {
	name string
	addr string
}

func (pil PeerInfoList) Readable() []string {
	readableList := make([]string, 0, len(pil))
	for _, peer := range pil {
		readableList = append(readableList, peer.PName()+"@"+peer.PAddr())
	}
	return readableList
}

func (lp LocalPeer) PAddr() string {
	return "local"
}

func (lp LocalPeer) Pick(key string) PeerInfo {
	return lp.Info()
}

func (lp LocalPeer) Info() PeerInfo {
	return LocalPeerInfo{
		name: "local",
		addr: "localhost",
	}
}

func (lp LocalPeer) Get(pi PeerInfo, key string) PeerResult {
	return PeerResult{Err: errors.New("not support")}
}

func (lp LocalPeer) Put(pi PeerInfo, key string, filename string, value []byte) PeerResult {
	return PeerResult{Err: errors.New("not support")}
}

func (lp LocalPeer) Delete(pi PeerInfo, key string) PeerResult {
	return PeerResult{Err: errors.New("not support")}
}

func (lp LocalPeer) PAdd(pis ...PeerInfo) {

}

func (lp LocalPeer) PDel(pis ...PeerInfo) {
}

func (lpi LocalPeerInfo) Equal(pi PeerInfo) bool {
	return lpi.name == pi.PName() && lpi.addr == pi.PAddr()
}

func (lpi LocalPeerInfo) PName() string {
	return lpi.name
}

func (lpi LocalPeerInfo) PAddr() string {
	return lpi.addr
}

func (lpi LocalPeerInfo) PStat() PeerStatType {
	return P_STAT_ONLINE
}

func (lpi LocalPeerInfo) Port() string {
	return "local"
}
