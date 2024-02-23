package node

import (
	"net"
	"strconv"
	"sync"

	"github.com/ciiim/syncmember"
)

type NodeService struct {
	self *Node

	sync *syncmember.SyncMember

	cMap *CMap

	onceRO sync.Once

	ro *NodeServiceRO
}

func NewNodeService(nodeName string, port int, replicas int) *NodeService {

	ns := &NodeService{
		sync:   syncmember.NewSyncMember(nodeName, syncmember.DefaultConfig().SetPort(port)),
		cMap:   NewCMap(replicas, nil),
		onceRO: sync.Once{},
	}
	self := ns.sync.Node()

	//FIXME: 临时解决方案
	addr := net.JoinHostPort(self.IP.String(), strconv.Itoa(self.Port+1))

	ns.self = NewNode(addr, self.Name)

	ns.sync.SetNodeDelegate(ns)

	ns.cMap.Add(ns.self)
	return ns
}

func (n *NodeService) Join(addr string) error {
	return n.sync.Join(addr)
}

func (n *NodeService) Self() *Node {
	return n.self
}

func (n *NodeService) Run() error {
	return n.sync.Run()
}

func (n *NodeService) Shutdown() {
	n.sync.Shutdown()
}

func (n *NodeService) AsyncRun() {
	go func() {
		_ = n.Run()
	}()
}
