package peers

type DefaultPeer struct{}

var _ Peer = (*DefaultPeer)(nil)

func (DefaultPeer) PName() string {
	return ""
}

func (DefaultPeer) PAddr() Addr {
	return nil
}

func (DefaultPeer) Pick(key string) PeerInfo {
	return nil
}

func (DefaultPeer) Info() PeerInfo {
	return nil
}

func (DefaultPeer) PAdd(pis ...PeerInfo) {

}

func (DefaultPeer) PDel(pis ...PeerInfo) {

}

func (DefaultPeer) PNext(key string) PeerInfo {
	return nil
}

func (DefaultPeer) PHandleSyncAction(pi PeerInfo, action PeerActionType) error {
	return nil
}

func (DefaultPeer) PActionTo(action PeerActionType, pi_to ...PeerInfo) error {
	return nil
}

func (DefaultPeer) GetPeerListFromPeer(pi PeerInfo) ([]PeerInfo, error) {
	return nil, nil
}

func (DefaultPeer) PList() []PeerInfo {
	return nil
}
