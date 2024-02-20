package node

import (
	"github.com/ciiim/cloudborad/node/chash"
)

// 并发安全
type CMap struct {
	*chash.ConsistentHash
}

// create a new consistent hash map
func NewCMap(replicas int, fn chash.ConsistentHashFn) *CMap {
	return &CMap{
		chash.NewConsistentHash(replicas, fn),
	}
}

func (c *CMap) Add(node *Node) {
	c.ConsistentHash.Add(node)
}

func (c *CMap) GetByNodeID(nodeID string) *Node {
	node, ok := c.ConsistentHash.GetByID(nodeID).(*Node)
	if !ok {
		return nil
	}
	return node
}

func (c *CMap) Get(key []byte) *Node {
	item := c.ConsistentHash.Get(key)
	if item == nil {
		return nil
	}
	node, ok := item.(*Node)
	if !ok {
		return nil
	}
	return node
}

func (c *CMap) Del(node *Node) {
	c.ConsistentHash.Del(node)
}

func (c *CMap) GetN(key []byte, n int) []*Node {
	nodes := make([]*Node, 0, n)
	for _, node := range c.ConsistentHash.GetN(key, n) {
		nodes = append(nodes, node.(*Node))
	}
	return nodes
}
