package node

import (
	"net"
	"strconv"

	"github.com/ciiim/syncmember"
)

var _ syncmember.NodeEventDelegate = (*NodeService)(nil)

func (n *NodeService) NotifyJoin(node *syncmember.Node) {
	addr := net.JoinHostPort(node.Addr().IP.String(), strconv.Itoa(node.Addr().Port+1))
	n.cMap.Add(NewNode(addr, node.Addr().Name))
}

func (n *NodeService) NotifyDead(node *syncmember.Node) {
	addr := net.JoinHostPort(node.Addr().IP.String(), strconv.Itoa(node.Addr().Port+1))
	n.cMap.Del(NewNode(addr, node.Addr().Name))
}

func (n *NodeService) NotifyAlive(node *syncmember.Node) {
	addr := net.JoinHostPort(node.Addr().IP.String(), strconv.Itoa(node.Addr().Port+1))
	n.cMap.Add(NewNode(addr, node.Addr().Name))
}
