package chash

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"

	"github.com/ciiim/cloudborad/internal/fs/peers"
	"golang.org/x/exp/slices"
)

type CHash func([]byte) uint32

// Consistent hash Map
type CMap struct {
	replicas      int
	hash          CHash
	realPeerInfos []peers.PeerInfo
	peerInfosHash []int

	rwmu sync.RWMutex

	hashMap sync.Map
}

// create a new consistent hash map
func NewCMap(replicas int, fn CHash) *CMap {
	m := &CMap{
		hash:     fn,
		hashMap:  sync.Map{},
		replicas: replicas,
		rwmu:     sync.RWMutex{},
	}
	if fn == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *CMap) Add(infos ...peers.PeerInfo) {
	var wg sync.WaitGroup
	wg.Add(len(infos))
	for _, info := range infos {
		go func(pi peers.PeerInfo) {
			m.rwmu.Lock()
			m.addRealNode(pi)
			for i := 0; i < m.replicas; i++ {
				hash := int(m.hash([]byte(strconv.Itoa(i) + pi.PName())))
				m.hashMap.Store(hash, pi)
				m.peerInfosHash = append(m.peerInfosHash, hash)
			}
			m.rwmu.Unlock()
			wg.Done()
		}(info)
	}
	wg.Wait()
	sort.Ints(m.peerInfosHash)
}

func (m *CMap) Del(infos ...peers.PeerInfo) {
	var wg sync.WaitGroup
	wg.Add(len(infos))
	for _, info := range infos {
		go func(pi peers.PeerInfo) {
			m.rwmu.Lock()
			m.delRealNode(pi)
			for i := 0; i < m.replicas; i++ {
				hash := int(m.hash([]byte(strconv.Itoa(i) + pi.PName())))
				m.hashMap.Delete(hash)
				for i, v := range m.peerInfosHash {
					if v == hash {
						m.peerInfosHash = append(m.peerInfosHash[:i], m.peerInfosHash[i+1:]...)
						break
					}
				}
			}
			m.rwmu.Unlock()
			wg.Done()
		}(info)
	}
	wg.Wait()
}

func (m *CMap) Get(key string) peers.PeerInfo {
	if len(m.peerInfosHash) == 0 {
		return nil
	}
	hash := int(m.hash([]byte(key)))
	index := sort.Search(len(m.peerInfosHash), func(i int) bool { return m.peerInfosHash[i] >= hash })
	m.rwmu.Lock()
	defer m.rwmu.Unlock()
	info, _ := m.hashMap.Load(m.peerInfosHash[index%len(m.peerInfosHash)])

	return info.(peers.PeerInfo)
}

func (m *CMap) addRealNode(info peers.PeerInfo) {
	m.realPeerInfos = append(m.realPeerInfos, info)
}

func (m *CMap) delRealNode(info peers.PeerInfo) {
	if idx := slices.Index[[]peers.PeerInfo](m.realPeerInfos, info); idx != -1 {
		m.realPeerInfos = slices.Delete[[]peers.PeerInfo](m.realPeerInfos, idx, idx+1)
	}
}

// Without virtual node
func (m *CMap) List() []peers.PeerInfo {
	infos := make([]peers.PeerInfo, len(m.realPeerInfos))
	m.rwmu.Lock()
	copy(infos, m.realPeerInfos)
	defer m.rwmu.Unlock()
	return infos
}

/*
When new peer added, some file needs to be moved to new peer.

So we need to get next peer to find the file.

You should incrase next if you cannot find the file in current peer.
*/
func (m *CMap) GetPeerNext(key string, next int) peers.PeerInfo {
	if len(m.peerInfosHash) == 0 {
		return nil
	}
	hash := int(m.hash([]byte(key)))
	index := sort.Search(len(m.peerInfosHash), func(i int) bool { return m.peerInfosHash[i] >= hash })
	m.rwmu.Lock()
	defer m.rwmu.Unlock()
	info, _ := m.hashMap.Load(m.peerInfosHash[index+next%len(m.peerInfosHash)])
	return info.(peers.PeerInfo)
}
