package node

import (
	"github.com/ciiim/syncmember"
)

var _ syncmember.NodeEventDelegate = (*NodeService)(nil)

func (n *NodeService) NotifyJoin(node *syncmember.Node) {
	n.cMap.Add(NewNode(node.Addr().String(), node.Addr().Name))
}

func (n *NodeService) NotifyDead(node *syncmember.Node) {
	n.cMap.Del(NewNode("", ""))
}

func (n *NodeService) NotifyAlive(node *syncmember.Node) {
	n.cMap.Add(NewNode(node.Addr().String(), node.Addr().Name))
}
