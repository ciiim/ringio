package peers

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/ciiim/cloudborad/internal/fs/peers"
)

type SpreadPeerList struct {
	peers.PeerInfo
}

type spread struct {
	pingInterval time.Duration
	r            *rand.Rand

	msgType reflect.Type
}

type gossip struct {
	msg any
}

type PeerList struct {
	peers   map[string]peers.PeerInfo
	version int64
	spread  *spread
}

func NewPeerList() *PeerList {
	spread := &spread{
		pingInterval: 1000 * time.Millisecond,
		r:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	pl := &PeerList{
		peers:  make(map[string]peers.PeerInfo),
		spread: spread,
	}
	return pl
}

func (s *spread) Start() {

}

func (s *spread) ping() {
}
func (s *spread) pong() {

}
