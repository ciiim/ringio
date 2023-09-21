// about peers
package peers

import "errors"

var (
	ErrPeerNotFound = errors.New("peer not found")
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

	// 节点下线
	P_ACTION_OFFLINE

	// 节点上线
	P_ACTION_ONLINE

	// 新节点接入集群
	P_ACTION_JOIN

	// 通知集群中其余节点有新节点加入
	P_ACTION_NEW

	// 节点退出集群 (主动退出)
	P_ACTION_QUIT
)

func (a PeerActionType) String() string {
	switch a {
	case P_ACTION_NONE:
		return "none"
	case P_ACTION_JOIN:
		return "join"
	case P_ACTION_NEW:
		return "new"
	case P_ACTION_QUIT:
		return "quit"
	default:
		return "unknown"
	}
}

type Addr interface {
	String() string
	IP() string
	Port() string
}
type Peer interface {
	PName() string
	PAddr() Addr
	Pick(key string) PeerInfo
	Info() PeerInfo
	PeerOperator
}

type PickPeerFn func(key string) PeerInfo

type PeerInfo interface {
	Equal(pi PeerInfo) bool
	PID() int64
	PName() string
	PAddr() Addr
	PVersion() int64
	PStat() PeerStatType
}

type PeerOperator interface {
	PAdd(pis ...PeerInfo)
	PDel(pis ...PeerInfo)
	PNext(key string) PeerInfo
	PHandleSyncAction(pi PeerInfo, action PeerActionType) error
	PActionTo(action PeerActionType, pi_to ...PeerInfo) error
	GetPeerListFromPeer(pi PeerInfo) ([]PeerInfo, error)
	PList() []PeerInfo
}
