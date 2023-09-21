package dfs

import (
	"log"
	"time"
)

type DPeerService struct {
	peers map[int64]*DPeer
}

func (p *DPeerService) BindPeer(peer *DPeer) {
	p.peers[peer.info.PeerID] = peer
}

func (p *DPeerService) pingAll() {
	for _, peer := range p.peers {
		peer.ping()
	}
}

func (p *DPeerService) RunSyncService(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			log.Println("tick")
			p.pingAll()
		}
	}
}
