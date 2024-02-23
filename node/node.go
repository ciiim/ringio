// about node
package node

import (
	"net"

	"github.com/ciiim/cloudborad/node/chash"
)

type Node struct {
	nodeID   string
	nodeIP   string
	nodePort string
	nodeName string
}

var _ chash.CHashItem = (*Node)(nil)

func NewNode(nodeAddr string, uniqueNodeName string) *Node {
	id := nodeAddr + uniqueNodeName
	addr, port, _ := net.SplitHostPort(nodeAddr)
	return &Node{
		nodeID:   id,
		nodeIP:   addr,
		nodePort: port,
		nodeName: uniqueNodeName,
	}
}

// return false if other is nil
func (n Node) Equal(other *Node) bool {
	if other == nil {
		return false
	}
	return n.ID() == other.ID()
}

func (n *Node) Compare(other chash.CHashItem) bool {
	return n.ID() == other.ID()
}

func (n Node) Name() string {
	return n.nodeName
}

func (n Node) Addr() string {
	return net.JoinHostPort(n.nodeIP, n.nodePort)
}

func (n Node) IP() string {
	return n.nodeIP
}

func (n Node) Port() string {
	return n.nodePort
}

func (n Node) ID() string {
	return n.nodeID
}
