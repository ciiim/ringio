package node

import (
	"github.com/ciiim/syncmember"
)

type NodeService struct {
	self *Node

	sync *syncmember.SyncMember

	cMap *CMap

	ro *NodeServiceRO
}

func NewNodeService(nodeName string, port int, replicas int) *NodeService {
	ns := &NodeService{
		sync: syncmember.NewSyncMember(nodeName, syncmember.DefaultConfig().SetPort(port)),
		cMap: NewCMap(replicas, nil),
	}
	self := ns.sync.Node()

	ns.self = NewNode(self.String(), self.Name)

	t := NewNode(ns.self.Addr(), ns.self.nodeName)

	ns.cMap.Add(t)
	return ns
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
		_ = n.sync.Run()
	}()
}
